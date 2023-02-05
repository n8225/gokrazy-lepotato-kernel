package utils

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"text/template"
)

func CreateTmpDir(n string) string {
	// We explicitly use /tmp, because Docker only allows volume mounts under
	// certain paths on certain platforms, see
	// e.g. https://docs.docker.com/docker-for-mac/osxfs/#namespaces for macOS.
	tmp, err := os.MkdirTemp("/tmp", "gokr-rebuild-"+n)
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	return tmp
}

func RunCompile(n, execName, tmp string) {
	log.Printf("compiling %s", n)
	var dockerRun *exec.Cmd
	v := map[string]string{"podman": "--userns=keep-id", "docker": ""}
	dockerRun = exec.Command(execName,
		"run",
		v[execName],
		"--rm",
		"--volume", tmp+":/tmp/buildresult:Z",
		"gokr-rebuild-"+n)
	dockerRun.Dir = tmp
	dockerRun.Stdout = os.Stdout
	dockerRun.Stderr = os.Stderr
	if err := dockerRun.Run(); err != nil {
		log.Fatalf("%s run: %v (cmd: %v)", execName, err, dockerRun.Args)
	}
}

func BuildGoBinary(n, tmp, pkgPath string) {
	log.Printf("building %s binary", n)
	cmd := exec.Command("go", "build", "-o", filepath.Join(tmp, "gokr-build-"+n), filepath.Join(pkgPath, "cmd/gokr-build-"+n))
	cmd.Env = append(os.Environ(), "GOOS=linux", "CGO_ENABLED=0")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("%v: %v", cmd.Args, err)
	}
}

func CreateContainer(n, pkgPath, tmp, execName string, buildFiles []string) {
	log.Printf("building %s container for %s compilation", execName, n)
	u := GetCurrentUser()

	dockerFileTmpl := template.Must(template.New("Dockerfile.tmpl").
		ParseFiles(filepath.Join(pkgPath, "Dockerfile.tmpl")))

	dockerFile, err := os.Create(filepath.Join(tmp, "Dockerfile"))
	if err != nil {
		log.Fatal(err)
	}

	if err := dockerFileTmpl.Execute(dockerFile, struct {
		Uid    string
		Gid    string
		Files  []string
		Binary string
	}{
		Uid:    u.Uid,
		Gid:    u.Gid,
		Files:  PrepareFiles(buildFiles, pkgPath, tmp),
		Binary: "gokr-build-" + n,
	}); err != nil {
		log.Fatal(err)
	}

	if err := dockerFile.Close(); err != nil {
		log.Fatal(err)
	}

	dockerBuild := exec.Command(execName,
		"build",
		"--rm=true",
		"--tag=gokr-rebuild-"+n,
		".")
	dockerBuild.Dir = tmp
	dockerBuild.Stdout = os.Stdout
	dockerBuild.Stderr = os.Stderr
	if err := dockerBuild.Run(); err != nil {
		log.Fatalf("%s build: %v (cmd: %v)", execName, err, dockerBuild.Args)
	}
}

func PrepareFiles(buildFiles []string, pkgPath, tmp string) []string {
	var files []string
	for _, f := range buildFiles {
		CopyFile(filepath.Join(pkgPath, f), filepath.Join(tmp, filepath.Base(f)))
		files = append(files, f)
	}
	return files
}

func GetCurrentUser() *user.User {
	u, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return u
}

func GetContainerExecutable() (string, error) {
	// Probe podman first, because the docker binary might actually
	// be a thin podman wrapper with podman behavior.
	choices := []string{"podman", "docker"}
	for _, exe := range choices {
		p, err := exec.LookPath(exe)
		if err != nil {
			continue
		}
		resolved, err := filepath.EvalSymlinks(p)
		if err != nil {
			return "", err
		}
		return resolved, nil
	}
	return "", fmt.Errorf("none of %v found in $PATH", choices)
}

func FindPkgDir() (string, error) {
	_, f, _, ok := runtime.Caller(1)
	if !ok {
		return "", errors.New("error get pkg directory")
	}
	return filepath.Join(filepath.Dir(f), "../../"), nil
}

// Download downloads a url
func Download(f string) error {
	log.Printf("Downloading %s to %s", f, filepath.Base(f))
	out, err := os.Create(filepath.Base(f))
	if err != nil {
		return err
	}
	defer out.Close()
	resp, err := http.Get(f)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if got, want := resp.StatusCode, http.StatusOK; got != want {
		return fmt.Errorf("unexpected HTTP status code for %s: got %d, want %d", f, got, want)
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}
	return out.Close()
}

// CopyFile copies a file from src to destination
func CopyFile(src, dest string) error {
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	st, err := in.Stat()
	if err != nil {
		return err
	}
	if err := out.Chmod(st.Mode()); err != nil {
		return err
	}
	return out.Close()
}

// Unzip unzips a zipped file
func Unzip(f string) {
	log.Printf("Unzipping %s to %s", f, filepath.Base(f))
	unzip := exec.Command("unzip", "-q", filepath.Base(f))
	unzip.Stdout = os.Stdout
	unzip.Stderr = os.Stderr
	if err := unzip.Run(); err != nil {
		log.Fatalf("unzip: %v", err)
	}
}

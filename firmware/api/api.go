package api

import (
    "log"
    "strconv"
    "strings"

    "fmt"
    "path"

    "github.com/ilya-pirogov/sonarmap/firmware/config"
    sonarmap "github.com/ilya-pirogov/sonarmap/map/config"
    "github.com/ilya-pirogov/sonarmap/firmware/shells"
)

type Api struct {
    ssh shells.TelnetShell
}

type AssetFile struct {
    Data []byte
    Name string
}

func New(ssh shells.TelnetShell) Api {
    return Api{
        ssh: ssh,
    }
}

func (a *Api) GetVersion() (ver int64) {
    var (
        out string
        err error
    )

    if out, err = a.ssh.Run("cat " + config.VerFile); err != nil { return 0 }
    lines := strings.Split(out, "\n")
    verStr := strings.TrimSpace(lines[len(lines) - 2])
    if ver, err = strconv.ParseInt(verStr, 10, 64); err != nil { return 0 }
    log.Printf("Version: %d", ver)
    return
}

func (a *Api) SetVersion(ver int64) error {
    _, err := a.ssh.Run("echo " + strconv.Itoa(int(ver)) + " > " + config.VerFile)
    return err
}

func (a *Api) StopService() error {
    _, err := a.ssh.Run("killall sonarmap")
    return err
}

func (a *Api) UploadSonarMap(data []byte) (err error) {
    log.Println("Start uploading sonarmap...")
    if _, err = a.ssh.Run("mount -o remount,rw /usr"); err != nil { return }
    log.Println("/usr remounted to read-write")
    if err = a.ssh.CopyBytes(data, config.DstSonarMap, "0755"); err != nil { return }
    log.Println("sonarmap copied to " + config.DstSonarMap)
    if _, err = a.ssh.Run("chmod +x " + config.DstSonarMap); err != nil { return }
    if _, err = a.ssh.Run("mount -o remount,ro /usr"); err != nil { return }
    log.Println("/usr remounted to read-only")
    return
}

func (a *Api) UploadWallpaper(data []byte) (err error) {
    log.Println("Start uploading wallpaper...")
    //if _, err = a.ssh.Run("mount -o remount,rw /usr"); err != nil { return }
    //log.Println("/usr remounted to read-write")
    if err = a.ssh.CopyBytes(data, sonarmap.Current.FileWallpaper, "0644"); err != nil { return }
    log.Println("wallpaper copied to " + sonarmap.Current.FileWallpaper)
    return
}

func (a *Api) UploadTranslations(assets []AssetFile) (err error) {
    log.Println("Start uploading translations...")
    if _, err = a.ssh.Run("mount -o remount,rw /usr"); err != nil { return }
    log.Println("/usr remounted to read-write")

    if _, err = a.ssh.Run(fmt.Sprintf("rm -rf %s.old", config.TranslationsDir)); err != nil { return }
    if _, err = a.ssh.Run(fmt.Sprintf("mv %s %s.old", config.TranslationsDir, config.TranslationsDir)); err != nil { return }
    if _, err = a.ssh.Run(fmt.Sprintf("mkdir %s", config.TranslationsDir)); err != nil { return }
    for _, asset := range assets {
        if err = a.ssh.CopyBytes(asset.Data, path.Join(config.TranslationsDir, asset.Name), "0755"); err != nil { return }
        log.Printf("%s copied to %s", asset.Name, path.Join(config.TranslationsDir, asset.Name))
    }
    if _, err = a.ssh.Run(fmt.Sprintf("chmod a+r %s", config.TranslationsDir)); err != nil { return }
    if _, err = a.ssh.Run(fmt.Sprintf("chmod -R a+r %s", config.TranslationsDir)); err != nil { return }

    if _, err = a.ssh.Run("mount -o remount,ro /usr"); err != nil { return }
    log.Println("/usr remounted to read-only")
    return
}

func (a *Api) PatchRc() (err error) {
    var (
        out string
    )
    log.Println("Start patching RC...")
    log.Println("Applying patch")
    if err = a.ssh.CopyBytes([]byte(config.RcPatch), "/tmp/rc.patch", "0644"); err != nil { return }
    if out, err = a.ssh.Run("patch -p0 -i /tmp/rc.patch"); err != nil { return }
    log.Println(out)
    //if out, err = a.ssh.Run("rm /tmp/rc.patch"); err != nil { return }

    return nil
}

func (a *Api) ChangePassword(newPassword string) (err error) {
    //log.Println("Changing password...")
    //if _, err = a.ssh.Run("echo \"" + newPassword + "\" | passwd --stdin root"); err != nil { return }

    return nil
}

func (a *Api) PowerOff() (err error) {
    log.Println("PowerOff...")
    if _, err = a.ssh.Run("poweroff"); err != nil { return }

    return nil
}

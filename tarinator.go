package tarinator

import(
    "archive/tar"
    "os"
    "io"
    "log"
    "path/filepath"
    "strings"
    "compress/gzip"
)

func TarGzFromFiles(paths []string, tarPath string) error {
    // set up the output file
    file, err := os.Create(tarPath)
    if err != nil {
        return err
    }

    defer file.Close()
     //set up the gzip writer
    gw := gzip.NewWriter(file)
    defer gw.Close()

    tw := tar.NewWriter(gw)
    defer tw.Close()

    // add each file as needed into the current tar archive
    for _,i := range paths {
        if err := tarit(i, "", tw); err != nil {
            return err
        }
    }

    return nil
}

func tarit(source, target string, tw *tar.Writer) error {
    info, err := os.Stat(source)
    if err != nil {
        return nil
    }

    var baseDir string
    if info.IsDir() {
        baseDir = filepath.Base(source)
    }

    return filepath.Walk(source,
        func(path string, info os.FileInfo, err error) error {
            if err != nil {
                return err
            }
            header, err := tar.FileInfoHeader(info, info.Name())
            if err != nil {
                return err
            }

            if baseDir != "" {
                header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
            }

            if err := tw.WriteHeader(header); err != nil {
                return err
            }

            if info.IsDir() {
                return nil
            }

            file, err := os.Open(path)
            if err != nil {
                return err
            }
            defer file.Close()
            _, err = io.Copy(tw, file)
            return err
        })
}

func untarit(extractPath, sourcefile string) error {
    file, err := os.Open(sourcefile)

    if err != nil {
        return err
    }

    defer file.Close()

    var fileReader io.ReadCloser = file

    // just in case we are reading a tar.gz file, add a filter to handle gzipped file
    if strings.HasSuffix(sourcefile, ".gz") {
        if fileReader, err = gzip.NewReader(file); err != nil {
            return err
        }
        defer fileReader.Close()
    }

    tarBallReader := tar.NewReader(fileReader)

    // Extracting tarred files
    for {
        header, err := tarBallReader.Next()
        if err != nil {
            if err == io.EOF {
                break
            }
            return err
        }

        // get the individual filename and extract to the current directory
        filename := filepath.Join(extractPath, header.Name)

        switch header.Typeflag {
        case tar.TypeDir:
            // handle directory
            err = os.MkdirAll(filename, os.FileMode(header.Mode)) // or use 0755 if you prefer

            if err != nil {
                return err
            }

        case tar.TypeReg:
            // handle normal file
            writer, err := os.Create(filename)

            if err != nil {
                return err
            }

            io.Copy(writer, tarBallReader)

            err = os.Chmod(filename, os.FileMode(header.Mode))

            if err != nil {
                return err
            }

            writer.Close()
        default:
            log.Printf("Unable to untar type : %c in file %s", header.Typeflag, filename)
        }
    }
    return nil
}

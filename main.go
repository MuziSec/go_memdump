package main

import (
	"fmt"
	"strings"
	"os/exec"
	"os"
	"strconv"
	"math"
	"io/ioutil"
	"io"
        "log"
	"archive/zip"
)

// ZipFiles compresses one or many files into a single zip archive file.
// Param 1: filename is the output zip file's name.
// Param 2: files is a list of files to add to the zip.
func ZipFiles(filename string, files []string) error {

    newZipFile, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer newZipFile.Close()

    zipWriter := zip.NewWriter(newZipFile)
    defer zipWriter.Close()

    // Add files to zip
    for _, file := range files {
        if err = AddFileToZip(zipWriter, file); err != nil {
            return err
        }
    }
    return nil
}

func AddFileToZip(zipWriter *zip.Writer, filename string) error {

    fileToZip, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer fileToZip.Close()

    // Get the file information
    info, err := fileToZip.Stat()
    if err != nil {
        return err
    }

    header, err := zip.FileInfoHeader(info)
    if err != nil {
        return err
    }

    // Using FileInfoHeader() above only uses the basename of the file. If we want
    // to preserve the folder structure we can overwrite this with the full path.
    header.Name = filename

    // Change to deflate to gain better compression
    // see http://golang.org/pkg/archive/zip/#pkg-constants
    header.Method = zip.Deflate

    writer, err := zipWriter.CreateHeader(header)
    if err != nil {
        return err
    }
    _, err = io.Copy(writer, fileToZip)
    return err
}

func chunk_file(file_to_chunk string) []string {

        file, err := os.Open(file_to_chunk)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer file.Close()

	fileInfo, _ := file.Stat()

	var fileSize int64 = fileInfo.Size()

	const fileChunk = 1 * (1000 << 20) // 1 GB, change if needed

	totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))

	fmt.Println("Chunking Memdump into 1 GB parts.")
        chunk_file_list := []string{}
	for i := uint64(0); i < totalPartsNum; i++ {

                 partSize := int(math.Min(fileChunk, float64(fileSize-int64(i*fileChunk))))
                 partBuffer := make([]byte, partSize)

                 file.Read(partBuffer)

                 // write to disk
                 fileName := "memchunk_" + strconv.FormatUint(i, 10)
		 chunk_file_list = append(chunk_file_list, fileName)
                 _, err := os.Create(fileName)

                 if err != nil {
                         fmt.Println(err)
                         os.Exit(1)
                 }

                 // write/save buffer to disk
                 ioutil.WriteFile(fileName, partBuffer, os.ModeAppend)

                 fmt.Println("Split to : ", fileName)
         }
         
	 return chunk_file_list

 }

func capture_windows_mem() string {

	fmt.Println("Starting Memory Capture Process")
        
	// Get System Arch
	arch_output, err := exec.Command("powershell.exe", "(Get-WmiObject Win32_OperatingSystem).OSArchitecture").Output()
        if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(arch_output))  

	// Capture memory based on Arch
	if strings.Contains(string(arch_output), "32-bit") {
		cmd := exec.Command("winpmem_mini_x86.exe", "memdump.raw")
		cmd.Start()
		cmd.Wait()
	} else if strings.Contains(string(arch_output), "64-bit") {
		cmd := exec.Command("winpmem_mini_x64_rc2.exe", "memdump.raw")
		cmd.Start()
		cmd.Wait()
	} else{
		fmt.Println("Error, couldn't get Windows OS Arch")
        }
	return "memdump.raw"
}

func capture_linux_mem() string {

	fmt.Println("Starting Memory Capture Process")
        
	memdump := "memdump.lime"
	
	// Make sure avml is executable
        cmd := exec.Command("chmod", "+x", "avml")
	cmd.Start()
	cmd.Wait()
        
	// Run avml
	cmd1 := exec.Command("./avml", memdump)
        cmd1.Start()
	cmd1.Wait()

	return memdump
}


func main() {
        
        if len(os.Args) < 2 {
		fmt.Println("Please enter Windows, Mac or Linux as an argument")
                fmt.Println("Exiting")
		os.Exit(3)
	}


	if os.Args[1] == "Windows" {
	    // Capture Windows Memory
	    fmt.Println("Capture Windows Memory")
	    file_to_chunk := capture_windows_mem()
  	    // Chunk the memory into 1 GB files
	    chunk_list := chunk_file(file_to_chunk)
	    // Zip the chunks up
	    output := "memdump.zip"
	    fmt.Println("Zipping chunks to memdump.zip")
	    if err := ZipFiles(output, chunk_list); err != nil {
		    panic(err)
	    }

	    fmt.Println("Zipped File:", output)
		
	    fmt.Println("Removing original memdump")
	    // Remove Memdump
	    e := os.Remove(file_to_chunk)
	    if e != nil {
		    log.Fatal(e)
	    }
	} else if os.Args[1] == "Linux" {
        	// Capture Linux Memory
		fmt.Println("Capture Linux Memory")
		file_to_chunk := capture_linux_mem()
  	        // Chunk the memory into 1 GB files
	        chunk_list := chunk_file(file_to_chunk)
	        // Zip the chunks up
	        output := "memdump.zip"
	        fmt.Println("Zipping chunks to memdump.zip")
	        if err := ZipFiles(output, chunk_list); err != nil {
		        panic(err)
	        }
	        fmt.Println("Zipped File:", output)
		
	        fmt.Println("Removing original memdump")
	        // Remove Memdump
	        e := os.Remove(file_to_chunk)
	        if e != nil {
		        log.Fatal(e)
	        }

	} else {
		fmt.Println("Unknown operating system argument.")
	}
        
        fmt.Println("Memdump Complete.")

}

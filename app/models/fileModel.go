package models

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/AdRoll/goamz/s3"
	"github.com/go-gorp/gorp"
	"github.com/irlTopper/lifevault/app/modules"
	"github.com/irlTopper/lifevault/app/modules/aws"
	"github.com/irlTopper/lifevault/app/modules/utils"
	"github.com/irlTopper/lifevault/app/utility"
	"github.com/jmcvetta/randutil"
	"github.com/revel/revel"
)

var ImgExts = []string{".jpg", ".jpeg", ".png"}
var fakeFileDef = []byte("FAKE FILE I'm AFRAID OLD BEAN!")

type FileDb struct {
	Id int64 `json:"id" db:"id"`
	// Data fields - a-z
	OriginalFileName string        `json:"originalFileName" db:"originalName"`
	PixelHeight      int64         `json:"pixelHeight" db:"pixelHeight"`
	PixelWidth       int64         `json:"pixelWidth" db:"pixelWidth"`
	S3Path           string        `json:"-" db:"s3Path"`
	Size             int64         `json:"size" db:"size"`
	ThreadId         sql.NullInt64 `json:"threadId" db:"threads_id"`
	Type             string        `json:"type" db:"type"`
	URL              string        `json:"URL" db:"URL"`
	// State fields
	State           string    `json:"-" db:"state"`
	CreatedAt       time.Time `json:"-" db:"createdAt"` //Auto updated
	UpdatedAt       time.Time `json:"-" db:"updatedAt"` //Auto updated
	CreatedByUserId int64     `json:"-" db:"createdBy_users_id"`
	UpdatedByUserId int64     `json:"-" db:"updatedBy_users_id"`
}

// implement the PreInsert and PreUpdate hooks
func (f *FileDb) PreInsert(s gorp.SqlExecutor) error {
	f.CreatedAt = time.Now().UTC()
	f.UpdatedAt = f.CreatedAt
	return nil
}

func (f *FileDb) PreUpdate(s gorp.SqlExecutor) error {
	f.UpdatedAt = time.Now().UTC()
	return nil
}

type TempFileRef struct {
	FileName          string  // eg. "tempduck.jpg"
	FileDb            *FileDb // a reference to the final file that this ended up in the database as
	LocalFilePath     string  // The full final path to the file on disk eg. "/tmp/installations/4/tempduck.jpg"
	OriginalFieldName string  // The name of the field that this came from eg. "attachment-1"
	OriginalFileName  string  // eg. "duck.jpg"
	Size              int64   `json:"size" db:"size"`
}

func GetFileExtension(fieldName string) (extension string) {

	pos := strings.LastIndex(fieldName, ".")
	if pos == -1 {
		return
	}

	return strings.ToLower(fieldName[pos:])
}

func GetPostedFileBytes(f *multipart.FileHeader) (*[]byte, error) {
	// Get Multipart file
	file, err := f.Open()
	if err != nil {
		return nil, errors.New("Failed to open uploaded file")
	}

	// Read file into memory
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.New("Error reading uploaded file contents")
	}

	return &b, nil
}

//TODO: This fucntion is never used!
func del_SaveUploadedFileToTemp(fieldName string, validFileExtensions []string, imageMaxWidth uint, imageMaxHeight uint, rc *revel.Controller, session *Session) (tempFileRef TempFileRef, err error) {

	var fileName string = rc.Request.Form.Get(fieldName)
	var fileBytesPtr *[]byte

	var isTestFile bool = false
	if strings.Index(fileName, "test:") == 0 {
		isTestFile = true
	}

	// Get the fileName and bytes
	if !isTestFile {

		if rc.Request.MultipartForm == nil {
			return tempFileRef, errors.New("Field '" + fieldName + "' not found")
		}
		// Get the file header for the fieldName
		fileHeaderArr, ok := rc.Request.MultipartForm.File[fieldName]
		if !ok || len(fileHeaderArr) == 0 || fileHeaderArr == nil || len(fileHeaderArr) == 0 {
			return tempFileRef, errors.New("Field '" + fieldName + "' not found")
		}
		fileHeader := fileHeaderArr[0]
		fileName = fileHeader.Filename
		// Get the file into memory
		fileBytesPtr, err = GetPostedFileBytes(fileHeader)
		if err != nil {
			return tempFileRef, err
		}
	} else {

		// For fake files, just remove the "test:" prefix
		fileName = strings.Replace(fileName, "test:", "", 1)

		// Get the file into memory
		// 1. See if we can load this test file from "testfiles/"
		testFilePath := filepath.FromSlash(revel.BasePath + "/tests/testfiles/" + fileName)
		fileBytes, err := ioutil.ReadFile(testFilePath)
		if err != nil {
			revel.ERROR.Println("FAILED TO READ FILE")
			// 2. If we can't read the file, just create a fake file in memory
			fileBytesPtr = &fakeFileDef
		} else {
			fileBytesPtr = &fileBytes
		}
	}

	tempFileRef, err = getTempFileRef(fieldName, fileName, fileBytesPtr, validFileExtensions, imageMaxWidth, imageMaxHeight, rc, session)
	if err != nil {
		return tempFileRef, err
	}

	return tempFileRef, nil
}

// Save as SaveUploadedFileToTemp but can handle an array of files
type TempFileOptions struct {
	FieldName           string
	File                []byte
	ValidFileExtensions []string
	ImageMaxWidth       uint
	ImageMaxHeight      uint
	MaxFilesize         utils.ByteSize
}

func SaveUploadedBinaryFileToTemp(options TempFileOptions, rc *revel.Controller, session *Session) (tempFileRefs []TempFileRef, err error) {

	var fileName = "BinaryFile"
	var fileBytesPtr = &options.File

	// Now check if the file is too large
	if int(options.MaxFilesize) > 0 && len(*fileBytesPtr) > int(options.MaxFilesize) {
		return tempFileRefs, errors.New(fmt.Sprintf("Upload filesize (%s) greater than max allowed (%s)", utils.ByteSize(len(*fileBytesPtr)).String(), options.MaxFilesize.String()))
	}

	// Get a temp file ref our of it
	tempFileRef, err := getTempFileRef(options.FieldName, fileName, fileBytesPtr, options.ValidFileExtensions, options.ImageMaxWidth, options.ImageMaxHeight, rc, session)
	if err != nil {
		return tempFileRefs, err
	}
	tempFileRefs = append(tempFileRefs, tempFileRef)

	return tempFileRefs, nil
}

func SaveUploadedFilesToTemp(options TempFileOptions, rc *revel.Controller, session *Session) (tempFileRefs []TempFileRef, err error) {

	var fileName string
	var fileBytesPtr *[]byte

	if rc.Request.MultipartForm == nil {
		return tempFileRefs, errors.New("Field '" + options.FieldName + "' not found")
	}
	// Get the file header for the options.FieldName
	fileHeaderArr, ok := rc.Request.MultipartForm.File[options.FieldName]
	numFiles := len(fileHeaderArr)
	if !ok || numFiles == 0 || fileHeaderArr == nil || len(fileHeaderArr) == 0 {
		return tempFileRefs, errors.New("Field '" + options.FieldName + "' not found")
	}

	for i := 0; i < numFiles; i++ {
		fileHeader := fileHeaderArr[i]
		fileName = fileHeader.Filename

		var isTestFile bool = false
		if strings.Index(fileName, "test:") == 0 {
			isTestFile = true
		}

		// Get the fileName and bytes
		if !isTestFile {
			// Get the file into memory
			fileBytesPtr, err = GetPostedFileBytes(fileHeader)
			if err != nil {
				return tempFileRefs, err
			}

			// Now check if the file is too large
			if int(options.MaxFilesize) > 0 && len(*fileBytesPtr) > int(options.MaxFilesize) {
				return tempFileRefs, errors.New(fmt.Sprintf("Upload filesize (%s) greater than max allowed (%s)", utils.ByteSize(len(*fileBytesPtr)).String(), options.MaxFilesize.String()))
			}

		} else {
			// For fake files, just remove the "test:" prefix
			fileName = strings.Replace(fileName, "test:", "", 1)

			// Get the file into memory
			// 1. See if we can load this test file from "testfiles/"
			testFilePath := filepath.FromSlash(revel.BasePath + "/tests/testfiles/" + fileName)
			fileBytes, err := ioutil.ReadFile(testFilePath)
			if err != nil {
				revel.ERROR.Println("FAILED TO READ TEST FILE", testFilePath)
				// 2. If we can't read the file, just create a fake file in memory
				fileBytesPtr = &fakeFileDef
			} else {
				fileBytesPtr = &fileBytes
			}
		}

		// Get a temp file ref our of it
		tempFileRef, err := getTempFileRef(options.FieldName, fileName, fileBytesPtr, options.ValidFileExtensions, options.ImageMaxWidth, options.ImageMaxHeight, rc, session)
		if err != nil {
			return tempFileRefs, err
		}
		tempFileRefs = append(tempFileRefs, tempFileRef)
	}

	return tempFileRefs, nil
}

// Todo- extact these options out to a struct
func getTempFileRef(fieldName string, fileName string, fileBytesPtr *[]byte, validFileExtensions []string, imageMaxWidth uint, imageMaxHeight uint, rc *revel.Controller, session *Session) (tempFileRef TempFileRef, err error) {

	// Remember the original field name and file name
	tempFileRef.OriginalFieldName = fieldName
	tempFileRef.OriginalFileName = fileName

	// Get the extension
	extension := GetFileExtension(tempFileRef.OriginalFileName)

	// If an image is required, check the extension is correct
	var validFileExtension bool
	if len(validFileExtensions) == 0 {
		validFileExtension = true // We are accepting any file type
	}
	for _, ext := range validFileExtensions {
		if strings.ToLower(extension) == strings.ToLower(ext) {
			validFileExtension = true
		}
	}

	var isImageRequired bool
	for _, ext := range ImgExts {
		if strings.ToLower(extension) == strings.ToLower(ext) {
			isImageRequired = true
		}
	}

	if !validFileExtension {
		return tempFileRef, errors.New(fmt.Sprintf("File must be of type %s!", strings.Join(validFileExtensions, ",")))
	}

	// Resize the image first
	if isImageRequired && imageMaxWidth > 0 {
		fileBytesPtr, err = modules.ThumbnailImage(fileBytesPtr, imageMaxWidth, imageMaxHeight)
		if err != nil {
			return tempFileRef, err
		}
	}

	// Get the temp dir - make the storage path if it doens't already exist
	var savePath string = utility.TempFolder

	// Build a new unique file name with userId, timestamp & random chars + original extension, if any
	randChars, _ := randutil.AlphaString(5)
	userIdStr := "0"
	if session != nil {
		userIdStr = strconv.FormatInt(session.Id, 10)
	}
	tempFileRef.FileName = userIdStr
	t := time.Now().UTC()
	tempFileRef.FileName += "." + fmt.Sprintf("%s%03d", t.Format("20060102150405"), t.Nanosecond()/1e6)
	tempFileRef.FileName += "." + tempFileRef.FileName
	tempFileRef.FileName += randChars
	if extension != "" {
		tempFileRef.FileName += extension
	}

	// Build the final save path
	tempFileRef.LocalFilePath = filepath.FromSlash(savePath + "/" + tempFileRef.FileName)

	// Set the file size
	tempFileRef.Size = int64(len(*fileBytesPtr))

	// Save the file
	err = ioutil.WriteFile(tempFileRef.LocalFilePath, *fileBytesPtr, 0777)
	modules.CheckErr(err, "Problem writing file", "I/O", map[string]interface{}{})

	return tempFileRef, nil
}

func SaveUploadedFileToS3(fileType string, fieldName string, file []byte, validFileExtensions []string, imageMaxWidth uint, imageMaxHeight uint, customHeaders map[string][]string, rc *revel.Controller, session *Session) (*FileDb, error) {

	var tempFileRefs []TempFileRef
	var err error

	// First we save this file to the temp folder
	if file == nil {
		tempFileRefs, err = SaveUploadedFilesToTemp(TempFileOptions{
			FieldName:           fieldName,
			ValidFileExtensions: validFileExtensions,
			ImageMaxWidth:       imageMaxWidth,
			ImageMaxHeight:      imageMaxHeight,
		},
			rc,
			session,
		)
	} else {
		tempFileRefs, err = SaveUploadedBinaryFileToTemp(TempFileOptions{
			File:           file,
			ImageMaxWidth:  imageMaxWidth,
			ImageMaxHeight: imageMaxHeight,
		},
			rc,
			session,
		)
	}

	if err != nil {
		return nil, errors.New("Error getting temporary directory path: " + err.Error())
	}

	if len(tempFileRefs) != 1 {
		return nil, errors.New("Only one file should be uploaded")
	}

	// Then move it to S3
	return MoveTempFileRefToS3(fileType, tempFileRefs[0], 0 /* threadId*/, customHeaders, rc, session)
}

func MoveTempFileRefToS3(fileType string, tempFileRef TempFileRef, threadId int64, customHeaders map[string][]string, rc *revel.Controller, session *Session) (*FileDb, error) {

	// TODO IN production, files go direct to S3, do not pass Go, do not collect $200
	// In dev mode, files get saved to a virtual S3 folder at /public/s3/

	// Prepend the S3Path with installation folder:
	s3Path := "u/" + strconv.FormatInt(session.User.Id, 10) + "/" + fileType

	var URL string
	if revel.DevMode && !revel.Config.BoolDefault("s3.test", false) {
		// Move the file from temp to the file path
		fileFileFolder := filepath.FromSlash(revel.BasePath + "/frontend/public/s3/" + s3Path + "/")
		if err := os.MkdirAll(fileFileFolder, 0777); err != nil {
			return nil, errors.New("Error creating directory path: " + err.Error())
		}
		finalFilePath := fileFileFolder + tempFileRef.FileName
		revel.ERROR.Println("Renaming ", tempFileRef.LocalFilePath, " to ", finalFilePath)
		if err := os.Rename(tempFileRef.LocalFilePath, finalFilePath); err != nil {
			return nil, err
		}

		// Build the URL to this file
		ServerURL, _ := revel.Config.String("server.URL")
		URL = ServerURL + "/desk/public/s3/" + s3Path + "/" + tempFileRef.FileName
	} else {
		// Get the uploaded file into memory
		data, err := ioutil.ReadFile(tempFileRef.LocalFilePath)
		if err != nil {
			return nil, err
		}

		var bucket *s3.Bucket
		if revel.DevMode && revel.Config.BoolDefault("s3.test", false) {
			// The S3 testing flag is set, let's upload to S3 tw-desk-dev bucket
			bucket = aws.AWS.S3.DeskDev
		} else {
			bucket = aws.AWS.S3.Desk
		}

		opts := s3.Options{
			CacheControl: "public; max-age=31536000",
		}

		// If we have custom headers, set them
		if len(customHeaders) > 0 {
			for headerKey, headerValue := range customHeaders {
				switch headerKey {
				case "Content-Disposition":
					opts.ContentDisposition = headerValue[0]
				}
			}
		}

		permission := s3.AuthenticatedRead

		if fileType != "attachment" {
			permission = s3.PublicRead
		}

		if err := bucket.Put(
			s3Path+"/"+tempFileRef.FileName,
			data,
			mime.TypeByExtension(filepath.Ext(tempFileRef.FileName)),
			permission,
			opts,
		); err != nil {
			return nil, err
		}

		URL = "https://s3.amazonaws.com/" + bucket.Name + "/" + s3Path + "/" + tempFileRef.FileName
	}

	s3Path += "/" + tempFileRef.FileName

	FileDb := FileDb{
		OriginalFileName: tempFileRef.OriginalFileName,
		Size:             tempFileRef.Size,
		S3Path:           s3Path,
		URL:              URL,
		Type:             fileType,
		CreatedByUserId:  session.Id,
		UpdatedByUserId:  session.Id,
		State:            "active",
	}

	if threadId > 0 {
		FileDb.ThreadId.Int64 = threadId
		FileDb.ThreadId.Valid = true
	}

	// Okay, we should first see if a reference to this file already exists
	fileId, _ := modules.DB.SelectInt(rc, "SELECT id FROM files WHERE s3Path = ?", s3Path)

	if fileId == 0 {
		modules.DB.Insert(rc, &FileDb)
	} else {
		FileDb.Id = fileId
		modules.DB.Update(rc, &FileDb)
	}

	return &FileDb, nil
}

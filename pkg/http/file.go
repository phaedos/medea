package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"medea/pkg/database/models"
	"medea/pkg/service"
	"mime"
	"mime/multipart"
	"net/http"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"medea/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

var testingChunkRootPath *string

var ErrWrongRangeHeader = errors.New("http range header format error")

var ErrWrongHTTPRange = errors.New("wrong http range header, start must be less than end")

const binaryContentType = "application/octet-stream"

type fileCreateInput struct {
	Token     string  `form:"token" binding:"required"`
	Nonce     string  `form:"nonce" header:"X-Request-Nonce" binding:"required,min=32,max=48"`
	Path      string  `form:"path" binding:"required,max=1000"`
	Sign      *string `form:"sign" binding:"omitempty"`
	Hash      *string `form:"hash" binding:"omitempty"`
	Size      *int    `form:"size" binding:"omitempty"`
	Overwrite *bool   `form:"overwrite,default=0" binding:"omitempty"`
	Rename    *bool   `form:"rename,default=0" binding:"omitempty"`
	Append    *bool   `form:"append,default=0" binding:"omitempty"`
	Hidden    *bool   `form:"hidden,default=0" binding:"omitempty"`
}

type fileReadInput struct {
	Token         string  `form:"token" binding:"required"`
	FileUID       string  `form:"fileUid" binding:"required"`
	Nonce         *string `form:"nonce" header:"X-Request-Nonce" binding:"omitempty,min=32,max=48"`
	Sign          *string `form:"sign" binding:"omitempty"`
	OpenInBrowser bool    `form:"openInBrowser,default=0" binding:"omitempty"`
}

type fileUpdateInput struct {
	Token   string  `form:"token" binding:"required"`
	FileUID string  `form:"fileUid" binding:"required"`
	Nonce   string  `form:"nonce" header:"X-Request-Nonce" binding:"omitempty,min=32,max=48"`
	Sign    *string `form:"sign" binding:"omitempty"`
	Hidden  *int8   `form:"hidden" binding:"omitempty"`
	Path    *string `form:"path" binding:"required,max=1000"`
}

type fileDeleteInput struct {
	Token   string  `form:"token" binding:"required"`
	Nonce   string  `form:"nonce" header:"X-Request-Nonce" binding:"omitempty,min=32,max=48"`
	FileUID string  `form:"fileUid" binding:"required"`
	Force   bool    `form:"force,default=0"  binding:"omitempty"`
	Sign    *string `form:"sign" binding:"omitempty"`
}

type directoryListInput struct {
	Token  string  `form:"token" binding:"required"`
	Nonce  string  `form:"nonce" header:"X-Request-Nonce" binding:"required,min=32,max=48"`
	Sign   *string `form:"sign" binding:"omitempty"`
	SubDir *string `form:"subDir,default=/" binding:"omitempty"`
	Sort   *string `form:"sort,default=-type" binding:"omitempty"`
	Limit  *int    `form:"limit,default=10" binding:"omitempty,min=10,max=20"`
	Offset *int    `form:"offset,default=0" binding:"omitempty,min=0"`
}

func FileCreateHandler(ctx *gin.Context) {
	var (
		fh     *multipart.FileHeader
		err    error
		buf    = bytes.NewBuffer(nil)
		reader io.Reader

		code     = 400
		reErrors map[string][]string
		success  bool
		data     interface{}

		db            = ctx.MustGet("db").(*gorm.DB)
		ip            = ctx.ClientIP()
		input         = ctx.MustGet("inputParam").(*fileCreateInput)
		fileCreateSrv = &service.FileCreate{
			BaseService: service.BaseService{DB: db},
			IP:          &ip,
			Path:        input.Path,
			Token:       ctx.MustGet("token").(*models.Token),
		}

		fileCreateValue interface{}
	)

	defer func() {
		ctx.JSON(code, &Response{
			RequestID: ctx.GetInt64("requestId"),
			Success:   success,
			Errors:    reErrors,
			Data:      data,
		})
	}()

	if fh, err = ctx.FormFile("file"); err != nil {
		if err != http.ErrMissingFile {
			reErrors = generateErrors(err, "file")
			return
		}
	} else {
		if reader, err = fh.Open(); err != nil {
			reErrors = generateErrors(err, "file")
			return
		}

		if _, err = io.Copy(buf, reader); err != nil {
			reErrors = generateErrors(err, "")
			return
		}

		if buf.Len() > models.ChunkSize {
			reErrors = generateErrors(models.ErrChunkExceedLimit, "file")
			return
		}

		reader = buf
		if reErrors = validateHashAndSize(input, buf); reErrors != nil {
			return
		}
	}

	fileCreateSrv.Reader = reader
	setFileCreateSrv(input, fileCreateSrv)

	if err := fileCreateSrv.Validate(); !reflect.ValueOf(err).IsNil() {
		reErrors = generateErrors(err, "")
		return
	}

	if fileCreateValue, err = fileCreateSrv.Execute(context.Background()); err != nil {
		reErrors = generateErrors(err, "")
		return
	}

	if data, err = fileResp(fileCreateValue.(*models.File), db); err != nil {
		reErrors = generateErrors(err, "")
		return
	}

	code = 200
	success = true
}

func setFileCreateSrv(input *fileCreateInput, fileCreateSrv *service.FileCreate) {
	if input.Hidden != nil && *input.Hidden {
		fileCreateSrv.Hidden = 1
	}
	if input.Overwrite != nil && *input.Overwrite {
		fileCreateSrv.Overwrite = 1
	}
	if input.Append != nil && *input.Append {
		fileCreateSrv.Append = 1
	}
	if input.Rename != nil && *input.Rename {
		fileCreateSrv.Rename = 1
	}

	if isTesting {
		fileCreateSrv.RootPath = testingChunkRootPath
	}
}

func validateHashAndSize(input *fileCreateInput, buf *bytes.Buffer) (reErrors map[string][]string) {
	if input.Hash != nil || input.Size != nil {
		if input.Size != nil && buf.Len() != *input.Size {
			reErrors = generateErrors(errors.New("the size of file doesn't match"), "size")
		}
		if input.Hash != nil {
			var (
				h   string
				err error
			)
			if h, err = utils.Sha256Hash2String(buf.Bytes()); err != nil {
				reErrors = generateErrors(err, "")
			}
			if h != *input.Hash {
				reErrors = generateErrors(errors.New("the hash of file doesn't match"), "hash")
			}
		}
	}
	return
}

func FileInfoHandler(ctx *gin.Context) {
	var (
		ip               = ctx.ClientIP()
		db               = ctx.MustGet("db").(*gorm.DB)
		err              error
		file             *models.File
		token            = ctx.MustGet("token").(*models.Token)
		input            = ctx.MustGet("inputParam").(*fileReadInput)
		requestID        = ctx.GetInt64("requestId")
		fileReadSrv      *service.FileRead
		readerSeeker     io.ReadSeeker
		fileReadSrvValue interface{}
	)

	if file, err = models.FindFileByUID(input.FileUID, false, db); err != nil {
		ctx.JSON(400, &Response{
			RequestID: requestID,
			Success:   false,
			Errors:    generateErrors(err, "fileUid"),
		})
		return
	}

	fileReadSrv = &service.FileRead{
		BaseService: service.BaseService{DB: db},
		Token:       token,
		File:        file,
		IP:          &ip,
	}

	if isTesting {
		fileReadSrv.RootPath = testingChunkRootPath
	}

	if err = fileReadSrv.Validate(); !reflect.ValueOf(err).IsNil() {
		ctx.JSON(400, &Response{
			RequestID: requestID,
			Success:   false,
			Errors:    generateErrors(err, ""),
		})
		return
	}

	if fileReadSrvValue, err = fileReadSrv.Execute(context.Background()); err != nil {
		ctx.JSON(400, &Response{
			RequestID: requestID,
			Success:   false,
			Errors:    generateErrors(err, ""),
		})
		return
	}

	readerSeeker = fileReadSrvValue.(io.ReadSeeker)
	rangeHeader := ctx.Request.Header.Get("Range")
	if rangeHeader == "" {
		readAllContent(ctx, readerSeeker, file, input)
		return
	}

	headers := map[string]string{
		"ETag":                file.Object.Hash,
		"Accept-Ranges":       "bytes",
		"Content-Type":        binaryContentType,
		"Last-Modified":       file.UpdatedAt.Format(time.RFC1123),
		"Content-Disposition": fmt.Sprintf(`attachment; filename="%s"`, file.Name),
	}
	headers["Content-Length"] = strconv.Itoa(file.Size)
	ctx.Set("ignoreRespBody", true)

	for k, v := range headers {
		ctx.Header(k, v)
	}
	ctx.Writer.WriteHeaderNow()
}

func FileReadHandler(ctx *gin.Context) {
	var (
		ip               = ctx.ClientIP()
		db               = ctx.MustGet("db").(*gorm.DB)
		err              error
		file             *models.File
		token            = ctx.MustGet("token").(*models.Token)
		input            = ctx.MustGet("inputParam").(*fileReadInput)
		requestID        = ctx.GetInt64("requestId")
		fileReadSrv      *service.FileRead
		fileReaderSeeker io.ReadSeeker
		fileReadSrvValue interface{}
	)

	if file, err = models.FindFileByUID(input.FileUID, false, db); err != nil {
		ctx.JSON(400, &Response{
			RequestID: requestID,
			Success:   false,
			Errors:    generateErrors(err, "fileUid"),
		})
		return
	}

	fileReadSrv = &service.FileRead{
		BaseService: service.BaseService{DB: db},
		Token:       token,
		File:        file,
		IP:          &ip,
	}

	if isTesting {
		fileReadSrv.RootPath = testingChunkRootPath
	}

	if err = fileReadSrv.Validate(); !reflect.ValueOf(err).IsNil() {
		ctx.JSON(400, &Response{
			RequestID: requestID,
			Success:   false,
			Errors:    generateErrors(err, ""),
		})
		return
	}

	if fileReadSrvValue, err = fileReadSrv.Execute(context.Background()); err != nil {
		ctx.JSON(400, &Response{
			RequestID: requestID,
			Success:   false,
			Errors:    generateErrors(err, ""),
		})
		return
	}
	fileReaderSeeker = fileReadSrvValue.(io.ReadSeeker)
	rangeHeader := ctx.Request.Header.Get("Range")
	if rangeHeader == "" {
		readAllContent(ctx, fileReaderSeeker, file, input)
		return
	}
	rangeHeaderPattern := regexp.MustCompile(`^bytes=(?P<start>\d*)-(?P<end>\d*)$`)
	if !rangeHeaderPattern.Match([]byte(rangeHeader)) {
		ctx.JSON(400, &Response{
			RequestID: requestID,
			Success:   false,
			Errors:    generateErrors(ErrWrongRangeHeader, ""),
		})
		return
	}
	rangePosition := strings.TrimPrefix(rangeHeader, "bytes=")
	rangeStart := 0
	rangeEnd := file.Size
	if rangePosition == "-" {
		readAllContent(ctx, fileReaderSeeker, file, input)
		return
	} else if strings.HasPrefix(rangePosition, "-") {
		rangeEnd, _ = strconv.Atoi(strings.TrimPrefix(rangePosition, "-"))
	} else if strings.HasSuffix(rangePosition, "-") {
		rangeStart, _ = strconv.Atoi(strings.TrimSuffix(rangePosition, "-"))
	} else {
		rangePositionSplit := strings.Split(rangePosition, "-")
		rangeStart, _ = strconv.Atoi(rangePositionSplit[0])
		rangeEnd, _ = strconv.Atoi(rangePositionSplit[1])
	}

	if rangeStart > rangeEnd {
		ctx.JSON(400, &Response{
			RequestID: requestID,
			Success:   false,
			Errors:    generateErrors(ErrWrongHTTPRange, ""),
		})
		return
	}
	readRangeContent(ctx, fileReaderSeeker, file, input, rangeStart, rangeEnd)
}

func readAllContent(ctx *gin.Context, readerSeeker io.ReadSeeker, file *models.File, input *fileReadInput) {
	headers := map[string]string{
		"ETag":                file.Object.Hash,
		"Accept-Ranges":       "bytes",
		"Content-Type":        binaryContentType,
		"Last-Modified":       file.UpdatedAt.Format(time.RFC1123),
		"Content-Disposition": fmt.Sprintf(`attachment; filename="%s"`, file.Name),
	}
	headers["Content-Length"] = strconv.Itoa(file.Size)
	if contentType := mime.TypeByExtension(path.Ext(file.Name)); contentType != "" {
		headers["Content-Type"] = contentType
	}
	if input.OpenInBrowser {
		headers["Content-Disposition"] = fmt.Sprintf(`inline; filename="%s"`, file.Name)
	}
	ctx.Set("ignoreRespBody", true)
	ctx.DataFromReader(http.StatusOK, int64(file.Size), headers["Content-Type"], readerSeeker, headers)
}

func readRangeContent(ctx *gin.Context, readerSeeker io.ReadSeeker, file *models.File, input *fileReadInput, start, end int) {
	if _, err := readerSeeker.Seek(int64(start), io.SeekStart); err != nil {
		ctx.JSON(400, &Response{
			RequestID: ctx.GetInt64("requestId"),
			Success:   false,
			Errors:    generateErrors(err, ""),
		})
		return
	}
	limitSize := int64(end-start) + 1
	limitReader := io.LimitReader(readerSeeker, limitSize)
	ctx.Set("ignoreRespBody", true)
	ctx.DataFromReader(http.StatusPartialContent, limitSize, binaryContentType, limitReader, map[string]string{
		"Content-Range": fmt.Sprintf("%d-%d/%d", start, end, file.Size),
	})
}

var ErrInvalidSortTypes = errors.New("invalid sort types, only one of type, -type, name, -name, time and -time")

func DirectoryListHandler(ctx *gin.Context) {
	var (
		ip                    = ctx.ClientIP()
		db                    = ctx.MustGet("db").(*gorm.DB)
		err                   error
		token                 = ctx.MustGet("token").(*models.Token)
		input                 = ctx.MustGet("inputParam").(*directoryListInput)
		directoryListSrv      *service.DirectoryList
		directoryListSrvValue interface{}
		directoryListSrvResp  *service.DirectoryListResponse

		code     = 400
		reErrors map[string][]string
		success  bool
		data     interface{}
	)

	defer func() {
		ctx.JSON(code, &Response{
			RequestID: ctx.GetInt64("requestId"),
			Success:   success,
			Errors:    reErrors,
			Data:      data,
		})
	}()

	if err = validateSort(*input.Sort); err != nil {
		reErrors = generateErrors(err, "sort")
		return
	}

	directoryListSrv = &service.DirectoryList{
		BaseService: service.BaseService{DB: db},
		Token:       token,
		IP:          &ip,
		SubDir:      *input.SubDir,
		Sort:        *input.Sort,
		Offset:      *input.Offset,
		Limit:       *input.Limit,
	}

	if err = directoryListSrv.Validate(); !reflect.ValueOf(err).IsNil() {
		reErrors = generateErrors(err, "")
		return
	}

	if directoryListSrvValue, err = directoryListSrv.Execute(context.Background()); err != nil {
		reErrors = generateErrors(err, "")
		return
	}

	directoryListSrvResp = directoryListSrvValue.(*service.DirectoryListResponse)

	result := map[string]interface{}{
		"total": directoryListSrvResp.Total,
		"pages": directoryListSrvResp.Pages,
	}

	items := make([]map[string]interface{}, len(directoryListSrvResp.Files))
	for index, item := range directoryListSrvResp.Files {
		if items[index], err = fileResp(&item, db); err != nil {
			reErrors = generateErrors(err, "")
			return
		}
	}

	result["items"] = items
	data = result
	code = 200
	success = true
}

var preDefinedSortTypes = []string{"type", "-type", "name", "-name", "time", "-time"}

func validateSort(sort string) error {
	for _, s := range preDefinedSortTypes {
		if s == sort {
			return nil
		}
	}
	return ErrInvalidSortTypes
}

func FileUpdateHandler(ctx *gin.Context) {
	var (
		ip                 = ctx.ClientIP()
		db                 = ctx.MustGet("db").(*gorm.DB)
		err                error
		file               *models.File
		token              = ctx.MustGet("token").(*models.Token)
		input              = ctx.MustGet("inputParam").(*fileUpdateInput)
		fileUpdateSrv      *service.FileUpdate
		fileUpdateSrvValue interface{}

		code     = 400
		reErrors map[string][]string
		success  bool
		data     interface{}
	)

	defer func() {
		ctx.JSON(code, &Response{
			RequestID: ctx.GetInt64("requestId"),
			Success:   success,
			Errors:    reErrors,
			Data:      data,
		})
	}()

	if file, err = models.FindFileByUID(input.FileUID, false, db); err != nil {
		reErrors = generateErrors(err, "fileUid")
		return
	}

	fileUpdateSrv = &service.FileUpdate{
		BaseService: service.BaseService{
			DB: db,
		},
		Token:  token,
		File:   file,
		IP:     &ip,
		Hidden: input.Hidden,
		Path:   input.Path,
	}

	if isTesting {
		fileUpdateSrv.RootPath = testingChunkRootPath
	}

	if err = fileUpdateSrv.Validate(); !reflect.ValueOf(err).IsNil() {
		reErrors = generateErrors(err, "")
		return
	}

	if fileUpdateSrvValue, err = fileUpdateSrv.Execute(context.Background()); err != nil {
		reErrors = generateErrors(err, "")
		return
	}

	if data, err = fileResp(fileUpdateSrvValue.(*models.File), db); err != nil {
		reErrors = generateErrors(err, "")
		return
	}

	code = 200
	success = true
}

func FileDeleteHandler(ctx *gin.Context) {
	var (
		ip                 = ctx.ClientIP()
		db                 = ctx.MustGet("db").(*gorm.DB)
		err                error
		file               *models.File
		token              = ctx.MustGet("token").(*models.Token)
		input              = ctx.MustGet("inputParam").(*fileDeleteInput)
		requestID          = ctx.GetInt64("requestId")
		fileDeleteSrv      *service.FileDelete
		fileDeleteSrvValue interface{}

		code     = 400
		reErrors map[string][]string
		success  bool
		data     interface{}
	)

	defer func() {
		ctx.JSON(code, &Response{
			RequestID: requestID,
			Success:   success,
			Errors:    reErrors,
			Data:      data,
		})
	}()

	if file, err = models.FindFileByUID(input.FileUID, false, db); err != nil {
		reErrors = generateErrors(err, "fileUid")
		return
	}

	fileDeleteSrv = &service.FileDelete{
		BaseService: service.BaseService{
			DB: db,
		},
		Token: token,
		File:  file,
		Force: &input.Force,
		IP:    &ip,
	}

	if err = fileDeleteSrv.Validate(); !reflect.ValueOf(err).IsNil() {
		reErrors = generateErrors(err, "system")
		return
	}

	if fileDeleteSrvValue, err = fileDeleteSrv.Execute(context.Background()); err != nil {
		reErrors = generateErrors(err, "system")
		return
	}

	if data, err = fileResp(fileDeleteSrvValue.(*models.File), db); err != nil {
		reErrors = generateErrors(err, "system")
		return
	}
	code = 200
	success = true
}

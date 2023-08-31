package logger

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type CheckTimeToOpenNewFileFunc func(lastOpenFileTime *time.Time, isNeverOpenFile bool) (string, bool)

var OpenNewFileByByDateHour CheckTimeToOpenNewFileFunc = func(lastOpenFileTime *time.Time, isNeverOpenFile bool) (string, bool) {
	if isNeverOpenFile {
		return time.Now().Format(logName + ".01-02.log"), true
	}

	lastOpenYear, lastOpenMonth, lastOpenDay := lastOpenFileTime.Date()
	nowYear, nowMonth, nowDay := time.Now().Date()
	if lastOpenDay != nowDay || lastOpenMonth != nowMonth || lastOpenYear != nowYear {
		return time.Now().Format(logName + ".01-02.log"), true
	}

	return "", false
}

type FileLoggerWriter struct {
	fp                        *os.File
	baseDir                   string
	maxFileSize               int64
	lastCheckIsFullAt         int64
	isFileFull                bool
	checkFileFullIntervalSecs int64
	checkTimeToOpenNewFile    CheckTimeToOpenNewFileFunc
	openCurrentFileTime       *time.Time
	currentFileName           string
	bufCh                     chan []byte
	isFlushing                atomic.Bool
	flushSignCh               chan struct{}
	flushDoneSignCh           chan error
}

func NewFileLoggerWriter(baseDir string, maxFileSize int64, checkFileFullIntervalSecs int64, checkTimeToOpenNewFile CheckTimeToOpenNewFileFunc, bufChanLen uint32) *FileLoggerWriter {
	return &FileLoggerWriter{
		baseDir:                   strings.TrimRight(baseDir, "/"),
		maxFileSize:               maxFileSize,
		checkFileFullIntervalSecs: checkFileFullIntervalSecs,
		checkTimeToOpenNewFile:    checkTimeToOpenNewFile,
		bufCh:                     make(chan []byte, bufChanLen),
		flushSignCh:               make(chan struct{}),
		flushDoneSignCh:           make(chan error),
	}
}

func (w *FileLoggerWriter) checkFileIsFull() (bool, error) {
	if w.lastCheckIsFullAt != 0 && w.lastCheckIsFullAt+w.checkFileFullIntervalSecs < time.Now().Unix() {
		return w.isFileFull, nil
	}

	fileInfo, err := w.fp.Stat()
	if err != nil {
		return false, err
	}

	w.isFileFull = fileInfo.Size() >= w.maxFileSize
	w.lastCheckIsFullAt = time.Now().Unix()

	return w.isFileFull, nil
}

func (w *FileLoggerWriter) tryOpenNewFile() error {
	var err error
	fileName, ok := w.checkTimeToOpenNewFile(w.openCurrentFileTime, w.openCurrentFileTime == nil)
	if !ok {
		if w.fp == nil {
			return errors.New("get first file name failed")
		}

		return nil
	}

	if w.fp == nil {
		if _, err = os.Stat(w.baseDir); err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			if err = os.MkdirAll(w.baseDir, 0755); err != nil {
				return err
			}
		}
	}

	if w.fp, err = os.OpenFile(w.baseDir+"/"+fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0755); err != nil {
		return err
	}

	openFileTime := time.Now()
	w.openCurrentFileTime = &openFileTime
	w.isFileFull = false
	w.lastCheckIsFullAt = 0
	w.currentFileName = fileName

	return nil
}

func (w *FileLoggerWriter) Flush() error {
	w.isFlushing.Store(true)
	w.flushSignCh <- struct{}{}
	return <-w.flushDoneSignCh
}

func (w *FileLoggerWriter) finishFlush(err error) {
	w.isFlushing.Store(false)
	w.flushDoneSignCh <- err
}

func (w *FileLoggerWriter) isFlushingNow() bool {
	return w.isFlushing.Load()
}

func (w *FileLoggerWriter) Write(logContent string) {
	select {
	case w.bufCh <- []byte(logContent):
	default:
		// never blocking main thread
		fmt.Println("log content cached buf full, lost:" + logContent)
	}
}

func (w *FileLoggerWriter) Loop() error {
	doWriteMoreAsPossible := func(buf []byte) error {
		for {
			var moreBuf []byte
			select {
			case moreBuf = <-w.bufCh:
				buf = append(buf, moreBuf...)
			default:
			}

			if moreBuf == nil {
				break
			}
		}

		if len(buf) == 0 {
			return nil
		}

		if err := w.tryOpenNewFile(); err != nil {
			return err
		}

		if isFull, err := w.checkFileIsFull(); err != nil {
			return err
		} else if isFull {
			fmt.Printf("log file %s is overflow max size %d bytes.\n", w.currentFileName, w.maxFileSize)
			return nil
		}

		bufLen := len(buf)
		var totalWrittenBytes int
		for {
			n, err := w.fp.Write(buf[totalWrittenBytes:])
			if err != nil {
				return err
			}
			totalWrittenBytes += n
			if totalWrittenBytes >= bufLen {
				break
			}
		}

		return nil
	}

	for {
		select {
		case buf := <-w.bufCh:
			if err := doWriteMoreAsPossible(buf); err != nil {
				return err
			}
		case _ = <-w.flushSignCh:
			if err := doWriteMoreAsPossible([]byte{}); err != nil {
				w.finishFlush(err)
				break
			}
			if err := w.fp.Sync(); err != nil {
				w.finishFlush(err)
				break
			}
			w.finishFlush(nil)
		}
	}
}

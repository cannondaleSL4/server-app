package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/bcrypt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"time"
)

var KubeClient *kubernetes.Clientset
var KubeConfig *restclient.Config
var namespace string
var uid string
var hashCost int

func init() {
	var err error
	KubeConfig = createConfig()
	KubeClient = createClient()
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules,
		configOverrides)
	namespace, _, err = kubeConfig.Namespace()
	if err != nil {
		getLogger().Error(fmt.Sprintf("failed to getting namespace: %v", err))
		panic(err)
	}

	namespaces, err := KubeClient.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		getLogger().Error(fmt.Sprintf("failed to getting namespaces: %v", err))
		panic(err)
	}

	for _, localNamespace := range namespaces.Items {
		if localNamespace.Name == namespace {
			uid = string(localNamespace.UID)
		}
	}
	dayOfYear := time.Now().YearDay()

	hashCost = dayOfYear

	if hashCost == 0 {
		hashCost = bcrypt.DefaultCost
	}
}

func createClient() *kubernetes.Clientset {
	kubeclient, err := kubernetes.NewForConfig(KubeConfig)
	if err != nil {
		getLogger().Error(fmt.Sprintf("failed to create new kubeClient: %v", err))
		return nil
	}
	return kubeclient
}

func createConfig() *restclient.Config {
	config, err := config.GetConfig()
	if err != nil {
		getLogger().Error(fmt.Sprintf("failed to build KubeConfig: %v", err))
		return nil
	}
	return config
}

func getLogger() *zap.Logger {
	var logLevel zapcore.Level
	_ = logLevel.Set(os.Getenv("LOG_LEVEL"))
	atom := zap.NewAtomicLevelAt(logLevel)
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atom,
	))

	defer logger.Sync()
	return logger
}

//func Compare(hash string) bool {
//	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(uid))
//	if err != nil {
//		log.Println(err)
//		return false
//	}
//	return true
//}

//export Compare
func Compare(hash string) *C.char {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(uid))
	if err != nil {
		return C.CString("false")
	}
	return C.CString("true")
}

func main() {}

//GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -buildmode=c-shared -o server.dll run.go
//go build -buildmode=c-shared -o server.dll run.go

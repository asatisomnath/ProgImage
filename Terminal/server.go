package Terminal

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/asatisomnath/ProgImage/Connection"
	"github.com/asatisomnath/ProgImage/SimpleStorageService"
	"github.com/google/uuid"
	"github.com/minio/minio-go"
	"github.com/spf13/cobra"
)

var addr string
var bucketName string
var accessKey string
var secretKey string
var endpoint string
var secure *bool

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVarP(&addr, "addr", "a", ":9090", "Bind address")
	serverCmd.Flags().StringVarP(&bucketName, "bucketname", "b", "ProgImage", "Storage bucket name")
	serverCmd.Flags().StringVarP(&accessKey, "accesskey", "k", "minio", "Storage access key")
	serverCmd.Flags().StringVarP(&secretKey, "secretkey", "s", "miniostorage", "Storage secret key")
	serverCmd.Flags().StringVarP(&endpoint, "endpoint", "e", "", "Storage endpoint")
	secure = serverCmd.Flags().Bool("secure", true, "Secure storage eg TLS")
	err := serverCmd.MarkFlagRequired("endpoint")
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error()) // nolint: errcheck,gas
		os.Exit(1)
	}
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Runs an Convertors processing Connection server",
	Long:  "Runs an Convertors processing Connection server",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := minio.New(endpoint, accessKey, secretKey, *secure)
		if err != nil {
			return err
		}

		uuid.New()
		is := SimpleStorageService.NewImageService(bucketName, c, uuid.New)
		if err := is.EnsureBucket(); err != nil {
			fmt.Fprintf(os.Stdout, "error checking bucket exists: %+v\n", err) // nolint: gas,errcheck
		}
		ih := Connection.NewImageHandler(is)
		s := Connection.Server{
			ImageHandler: *ih,
			Addr:         addr,
		}

		done := make(chan bool)
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)

		go func() {
			<-quit
			fmt.Fprint(os.Stdout, "stopping server\n") // nolint: gas,errcheck

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			if err := s.Stop(ctx); err != nil {
				fmt.Fprintf(os.Stdout, "unable to shutdown gracefully %s\n:", err) // nolint: gas,errcheck
			}
			close(done)
		}()

		fmt.Fprintf(os.Stdout, "started server on %s\n", addr) // nolint: gas,errcheck
		if err := s.Start(os.Stdout); err != nil {
			return err
		}

		<-done
		fmt.Fprint(os.Stdout, "goodbye\n") // nolint: gas,errcheck

		return nil
	},
}

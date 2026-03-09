package storage

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/AndroidPoet/appwrite-cli/internal/api"
	"github.com/AndroidPoet/appwrite-cli/internal/cli"
	"github.com/AndroidPoet/appwrite-cli/internal/completion"
	"github.com/AndroidPoet/appwrite-cli/internal/output"
)

var (
	bucketID      string
	fileID        string
	name          string
	permissions   []string
	maxFileSize   int
	allowedExts   []string
	enabled       bool
	encryption    bool
	antivirus     bool
	compression   string
	limit         int
	offset        int
	allPages      bool
)

// StorageCmd manages storage
var StorageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Manage storage buckets and files",
	Long:  `List, create, update, and delete storage buckets and files.`,
}

// Bucket commands
var listBucketsCmd = &cobra.Command{
	Use:   "list-buckets",
	Short: "List all storage buckets",
	RunE:  runListBuckets,
}

var getBucketCmd = &cobra.Command{
	Use:   "get-bucket",
	Short: "Get bucket details",
	RunE:  runGetBucket,
}

var createBucketCmd = &cobra.Command{
	Use:   "create-bucket",
	Short: "Create a new storage bucket",
	RunE:  runCreateBucket,
}

var updateBucketCmd = &cobra.Command{
	Use:   "update-bucket",
	Short: "Update a storage bucket",
	RunE:  runUpdateBucket,
}

var deleteBucketCmd = &cobra.Command{
	Use:   "delete-bucket",
	Short: "Delete a storage bucket",
	RunE:  runDeleteBucket,
}

// File commands
var listFilesCmd = &cobra.Command{
	Use:   "list-files",
	Short: "List files in a bucket",
	RunE:  runListFiles,
}

var getFileCmd = &cobra.Command{
	Use:   "get-file",
	Short: "Get file details",
	RunE:  runGetFile,
}

var deleteFileCmd = &cobra.Command{
	Use:   "delete-file",
	Short: "Delete a file",
	RunE:  runDeleteFile,
}

func init() {
	listBucketsCmd.Flags().IntVar(&limit, "limit", 25, "number of results per page")
	listBucketsCmd.Flags().IntVar(&offset, "offset", 0, "offset for pagination")
	listBucketsCmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")

	getBucketCmd.Flags().StringVar(&bucketID, "bucket-id", "", "bucket ID")
	getBucketCmd.MarkFlagRequired("bucket-id")
	getBucketCmd.RegisterFlagCompletionFunc("bucket-id", completion.BucketIDs())

	createBucketCmd.Flags().StringVar(&bucketID, "bucket-id", "", "bucket ID (auto-generated if empty)")
	createBucketCmd.Flags().StringVar(&name, "name", "", "bucket name")
	createBucketCmd.Flags().StringSliceVar(&permissions, "permissions", nil, "bucket permissions")
	createBucketCmd.Flags().IntVar(&maxFileSize, "max-file-size", 30000000, "max file size in bytes")
	createBucketCmd.Flags().StringSliceVar(&allowedExts, "allowed-extensions", nil, "allowed file extensions")
	createBucketCmd.Flags().BoolVar(&enabled, "enabled", true, "enable the bucket")
	createBucketCmd.Flags().BoolVar(&encryption, "encryption", true, "enable encryption")
	createBucketCmd.Flags().BoolVar(&antivirus, "antivirus", true, "enable antivirus")
	createBucketCmd.Flags().StringVar(&compression, "compression", "none", "compression: none, gzip, zstd")
	createBucketCmd.MarkFlagRequired("name")

	updateBucketCmd.Flags().StringVar(&bucketID, "bucket-id", "", "bucket ID")
	updateBucketCmd.Flags().StringVar(&name, "name", "", "bucket name")
	updateBucketCmd.Flags().StringSliceVar(&permissions, "permissions", nil, "bucket permissions")
	updateBucketCmd.Flags().IntVar(&maxFileSize, "max-file-size", 30000000, "max file size in bytes")
	updateBucketCmd.Flags().StringSliceVar(&allowedExts, "allowed-extensions", nil, "allowed file extensions")
	updateBucketCmd.Flags().BoolVar(&enabled, "enabled", true, "enable the bucket")
	updateBucketCmd.Flags().BoolVar(&encryption, "encryption", true, "enable encryption")
	updateBucketCmd.Flags().BoolVar(&antivirus, "antivirus", true, "enable antivirus")
	updateBucketCmd.Flags().StringVar(&compression, "compression", "none", "compression: none, gzip, zstd")
	updateBucketCmd.MarkFlagRequired("bucket-id")
	updateBucketCmd.MarkFlagRequired("name")
	updateBucketCmd.RegisterFlagCompletionFunc("bucket-id", completion.BucketIDs())

	var confirm bool
	deleteBucketCmd.Flags().StringVar(&bucketID, "bucket-id", "", "bucket ID")
	deleteBucketCmd.Flags().BoolVar(&confirm, "confirm", false, "confirm deletion")
	deleteBucketCmd.MarkFlagRequired("bucket-id")
	deleteBucketCmd.RegisterFlagCompletionFunc("bucket-id", completion.BucketIDs())

	// File commands
	for _, cmd := range []*cobra.Command{listFilesCmd, getFileCmd, deleteFileCmd} {
		cmd.Flags().StringVar(&bucketID, "bucket-id", "", "bucket ID")
		cmd.MarkFlagRequired("bucket-id")
		cmd.RegisterFlagCompletionFunc("bucket-id", completion.BucketIDs())
	}

	listFilesCmd.Flags().IntVar(&limit, "limit", 25, "number of results per page")
	listFilesCmd.Flags().IntVar(&offset, "offset", 0, "offset for pagination")
	listFilesCmd.Flags().BoolVar(&allPages, "all", false, "fetch all pages")

	getFileCmd.Flags().StringVar(&fileID, "file-id", "", "file ID")
	getFileCmd.MarkFlagRequired("file-id")

	deleteFileCmd.Flags().StringVar(&fileID, "file-id", "", "file ID")
	deleteFileCmd.Flags().BoolVar(&confirm, "confirm", false, "confirm deletion")
	deleteFileCmd.MarkFlagRequired("file-id")

	StorageCmd.AddCommand(listBucketsCmd)
	StorageCmd.AddCommand(getBucketCmd)
	StorageCmd.AddCommand(createBucketCmd)
	StorageCmd.AddCommand(updateBucketCmd)
	StorageCmd.AddCommand(deleteBucketCmd)
	StorageCmd.AddCommand(listFilesCmd)
	StorageCmd.AddCommand(getFileCmd)
	StorageCmd.AddCommand(deleteFileCmd)
}

type BucketInfo struct {
	ID            string   `json:"$id"`
	Name          string   `json:"name"`
	Enabled       bool     `json:"enabled"`
	MaxFileSize   int      `json:"maximumFileSize"`
	AllowedExts   []string `json:"allowedFileExtensions"`
	Encryption    bool     `json:"encryption"`
	Antivirus     bool     `json:"antivirus"`
	Compression   string   `json:"compression"`
	CreatedAt     string   `json:"$createdAt,omitempty"`
}

type FileInfo struct {
	ID        string `json:"$id"`
	BucketID  string `json:"bucketId"`
	Name      string `json:"name"`
	Size      int    `json:"sizeOriginal"`
	MimeType  string `json:"mimeType"`
	CreatedAt string `json:"$createdAt,omitempty"`
}

func parseTimeout() time.Duration {
	d, err := time.ParseDuration(cli.GetTimeout())
	if err != nil {
		return 60 * time.Second
	}
	return d
}

func runListBuckets(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	if allPages {
		var items []BucketInfo
		err := client.ListAll(ctx, "/storage/buckets", limit, "buckets", func(raw json.RawMessage) error {
			var page []BucketInfo
			if err := json.Unmarshal(raw, &page); err != nil {
				return err
			}
			items = append(items, page...)
			return nil
		})
		if err != nil {
			return err
		}
		return output.Print(items)
	}

	var resp struct {
		Total   int          `json:"total"`
		Buckets []BucketInfo `json:"buckets"`
	}
	path := fmt.Sprintf("/storage/buckets?limit=%d&offset=%d", limit, offset)
	if err := client.Get(ctx, path, &resp); err != nil {
		return err
	}

	if resp.Total > offset+limit {
		output.PrintInfo("Showing %d of %d. Use --offset=%d for next page.", limit, resp.Total, offset+limit)
	}

	return output.Print(resp.Buckets)
}

func runGetBucket(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	var bucket BucketInfo
	if err := client.Get(ctx, "/storage/buckets/"+bucketID, &bucket); err != nil {
		return err
	}
	return output.Print(bucket)
}

func runCreateBucket(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would create bucket '%s'", name)
		return nil
	}

	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	body := map[string]interface{}{
		"name":                    name,
		"enabled":                 enabled,
		"maximumFileSize":         maxFileSize,
		"allowedFileExtensions":   allowedExts,
		"encryption":              encryption,
		"antivirus":               antivirus,
		"compression":             compression,
	}
	if bucketID != "" {
		body["bucketId"] = bucketID
	} else {
		body["bucketId"] = "unique()"
	}
	if len(permissions) > 0 {
		body["permissions"] = permissions
	}

	var bucket BucketInfo
	if err := client.Post(ctx, "/storage/buckets", body, &bucket); err != nil {
		return err
	}

	output.PrintSuccess("Bucket '%s' created (ID: %s)", name, bucket.ID)
	return output.Print(bucket)
}

func runUpdateBucket(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would update bucket '%s'", bucketID)
		return nil
	}

	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	body := map[string]interface{}{
		"name":                    name,
		"enabled":                 enabled,
		"maximumFileSize":         maxFileSize,
		"allowedFileExtensions":   allowedExts,
		"encryption":              encryption,
		"antivirus":               antivirus,
		"compression":             compression,
	}
	if len(permissions) > 0 {
		body["permissions"] = permissions
	}

	var bucket BucketInfo
	if err := client.Put(ctx, "/storage/buckets/"+bucketID, body, &bucket); err != nil {
		return err
	}

	output.PrintSuccess("Bucket '%s' updated", bucketID)
	return output.Print(bucket)
}

func runDeleteBucket(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	if err := cli.CheckConfirm(cmd); err != nil {
		return err
	}
	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would delete bucket '%s'", bucketID)
		return nil
	}

	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	if err := client.Delete(ctx, "/storage/buckets/"+bucketID); err != nil {
		return err
	}

	output.PrintSuccess("Bucket '%s' deleted", bucketID)
	return nil
}

func runListFiles(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	base := fmt.Sprintf("/storage/buckets/%s/files", bucketID)

	if allPages {
		var items []FileInfo
		err := client.ListAll(ctx, base, limit, "files", func(raw json.RawMessage) error {
			var page []FileInfo
			if err := json.Unmarshal(raw, &page); err != nil {
				return err
			}
			items = append(items, page...)
			return nil
		})
		if err != nil {
			return err
		}
		return output.Print(items)
	}

	var resp struct {
		Total int        `json:"total"`
		Files []FileInfo `json:"files"`
	}
	path := fmt.Sprintf("%s?limit=%d&offset=%d", base, limit, offset)
	if err := client.Get(ctx, path, &resp); err != nil {
		return err
	}

	if resp.Total > offset+limit {
		output.PrintInfo("Showing %d of %d. Use --offset=%d for next page.", limit, resp.Total, offset+limit)
	}

	return output.Print(resp.Files)
}

func runGetFile(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	var file FileInfo
	path := fmt.Sprintf("/storage/buckets/%s/files/%s", bucketID, fileID)
	if err := client.Get(ctx, path, &file); err != nil {
		return err
	}
	return output.Print(file)
}

func runDeleteFile(cmd *cobra.Command, args []string) error {
	if err := cli.RequireProject(cmd); err != nil {
		return err
	}
	if err := cli.CheckConfirm(cmd); err != nil {
		return err
	}
	if cli.IsDryRun() {
		output.PrintInfo("Dry run: would delete file '%s'", fileID)
		return nil
	}

	client, err := api.NewClient(cli.GetProjectID(), parseTimeout())
	if err != nil {
		return err
	}
	ctx, cancel := client.Context()
	defer cancel()

	path := fmt.Sprintf("/storage/buckets/%s/files/%s", bucketID, fileID)
	if err := client.Delete(ctx, path); err != nil {
		return err
	}

	output.PrintSuccess("File '%s' deleted", fileID)
	return nil
}

package mirroring

import (
	"context"
	minio "github.com/minio/minio/cmd"
	"github.com/minio/minio/pkg/hash"
	"io"
	"storj.io/ditto/pkg/config"
	l "storj.io/ditto/pkg/logger"
)

//MirroringObjectLayer is
type MirroringObjectLayer struct {
	minio.GatewayUnsupported
	Prime  minio.ObjectLayer
	Alter  minio.ObjectLayer
	Logger l.Logger
	Config *config.Config
}

//ObjectLayer interface---------------------------------------------------------------------------------------------------------------------

func (m *MirroringObjectLayer) Shutdown(ctx context.Context) error {
	return nil
}

func (m *MirroringObjectLayer) StorageInfo(ctx context.Context) (storageInfo minio.StorageInfo) {
	return storageInfo
}

func (m *MirroringObjectLayer) MakeBucketWithLocation(ctx context.Context, bucket string, location string) error {

	h := NewMakeBucketHandler(m, ctx, bucket, location)

	return h.Process()
}

// Returns bucket name and creation date of the bucket.
// Parameters:
// ctx    - current context.
// bucket - bucket name.
func (m *MirroringObjectLayer) GetBucketInfo(ctx context.Context, bucket string) (bucketInfo minio.BucketInfo, err error) {

	h := NewGetBucketInfoHandler(m, ctx, bucket)

	return h.Process()
}

// Returns a list_cmd of all buckets.
// Parameters:
// ctx - current context.
func (m *MirroringObjectLayer) ListBuckets(ctx context.Context) (buckets []minio.BucketInfo, err error) {

	h := NewListBucketsHandler(m, ctx)

	return h.Process()
}

// Deletes the bucket named in the URI.
// All objects (including all object versions and delete markers) in the bucket
// must be deleted before the bucket itself can be deleted.
// Parameters:
// ctx    - current context.
// bucket - bucket name.
func (m *MirroringObjectLayer) DeleteBucket(ctx context.Context, bucket string) error {

	h := NewDeleteBucketHandler(m, ctx, bucket)

	return h.Process()
}

// ListObjects is a paginated operation.
// Multiple API calls may be issued in order to retrieve the entire data set of results.
// You can disable pagination by providing the --no-paginate argument.
// Returns some or all (up to 1000) of the objects in a bucket.
// You can use the request parameters as selection criteria to return a subset of the objects in a bucket.
// Parameters:
// ctx       - current context.
// bucket    - bucket name.
// prefix    - Limits the response to keys that begin with the specified prefix.
// marker    - Specifies the key to start with when listing objects in a bucket.
// 			   Amazon S3 returns object keys in UTF-8 binary order, starting with key after the marker in order.
// delimiter - is a character you use to group keys.
// maxKeys   - Sets the maximum number of keys returned in the response body.
// 			   If you want to retrieve fewer than the default 1,000 keys, you can add this to your request.
//             Default value is 1000
func (m *MirroringObjectLayer) ListObjects(ctx context.Context,
										   bucket string,
										   prefix string,
										   marker string,
										   delimiter string,
										   maxKeys int) (minio.ListObjectsInfo, error) {

	h := NewListObjectsHandler(m, ctx,bucket, prefix, marker, delimiter, maxKeys)

	return h.Process()
}

// This implementation of the GET operation returns some or all (up to 1,000) of the objects in a bucket.
// You can use the request parameters as selection criteria to return a subset of the objects in a bucket.
// A 200 OK response can contain valid or invalid XML.
// Make sure to design your application to parse the contents of the response and handle it appropriately.
// Parameters:
// ctx 	   - current context.
// bucket  - bucket name.
// prefix  - Limits the response to keys that begin with the specified prefix.
// cntnTkn - when the response to this API call is truncated (that is, the IsTruncated response element value is true),
// 			 the response also includes the NextContinuationToken element.
// 			 To list_cmd the next set of objects, you can use the NextContinuationToken element in the next request as the continuation-token.
// 			 Amazon S3 returns object keys in UTF-8 binary order, starting with key after the marker in order.
// delim   - is a character you use to group keys.
// maxKeys - Sets the maximum number of keys returned in the response body.
// 		     If you want to retrieve fewer than the default 1,000 keys, you can add this to your request.
//           Default value is 1000.
func (m *MirroringObjectLayer) ListObjectsV2(ctx        context.Context,
											 bucket     string,
											 prefix     string,
											 cntnTkn    string,
											 delim      string,
											 maxKeys    int,
											 fetchOwner bool,
											 startAfter string) (minio.ListObjectsV2Info, error) {

	h := NewListObjectsV2Handler(m, ctx, bucket, prefix, cntnTkn, delim, startAfter, maxKeys, fetchOwner)

	return h.Process()
}

// Retrieves an object
// Parameters:
// ctx         - current context.
// bucket      - bucket name.
// object      - object name.
// startOffset - indicates the starting read location of the object.
// length      - indicates the total length of the object.
// etag        - An ETag is an opaque identifier assigned by a web server
// 			     to a specific version of a resource found at a URL
// opts        -
func (m *MirroringObjectLayer) GetObject(ctx 		 context.Context,
										 bucket 	 string,
										 object      string,
										 startOffset int64,
										 length 	 int64,
										 writer 	 io.Writer,
									     etag 	     string,
										 opts 		 minio.ObjectOptions) (err error) {

	h := newGetHandler(m.Prime, m.Alter, false)
	return h.process(ctx, bucket, object, startOffset, length, writer, etag, opts)
}

// Returns information about object.
// Parameters:
// ctx    - current context.
// bucket - bucket name.
// object - object name.
func (m *MirroringObjectLayer) GetObjectInfo(ctx    context.Context,
											 bucket string,
											 object string,
											 opts   minio.ObjectOptions) (objInfo minio.ObjectInfo, err error) {

	h := NewGetObjectInfoHandler(m, ctx, bucket, object, opts)

	return h.Process()
}

// PutObject adds an object to a bucket.
// Parameters:
// ctx         - current context.
// bucket      - bucket name.
// object      - object name.
// metadata    - A map of metadata to store with the object.
func (m *MirroringObjectLayer) PutObject(ctx context.Context, bucket string, object string, data *hash.Reader, metadata map[string]string, opts minio.ObjectOptions) (objInfo minio.ObjectInfo, err error) {
	//TODO: decide prime and alter based on config
	h := newPutHandler(m.Prime, m.Alter, m.Logger)
	return h.process(ctx, bucket, object, data, metadata, opts)
}

// Creates a cp of an object that is already stored in a bucket.
// Parameters:
// srcBucket  - The name of the source bucket
// srcObject  - Key name of the source object
// destBucket - The name of the destination bucket
// destObject - Key name of the destination object
// srcInfo    - represents object metadata
func (m *MirroringObjectLayer) CopyObject(ctx 		 context.Context,
										  srcBucket  string,
										  srcObject  string,
										  destBucket string,
										  destObject string,
										  srcInfo 	 minio.ObjectInfo,
										  srcOpts 	 minio.ObjectOptions,
										  destOpts 	 minio.ObjectOptions) (minio.ObjectInfo, error) {

	h := NewCopyObjectHandler(m, ctx, srcBucket, srcObject, destBucket, destObject, srcInfo, srcOpts, destOpts)

	return h.Process()
}

// Deletes the bucket named in the URI.
// Parameters:
// ctx    - current context.
// bucket - bucket name.
// object - object name
func (m *MirroringObjectLayer) DeleteObject(ctx context.Context, bucket, object string) error {

	h := NewDeleteObjectHandler(m, ctx, bucket, object)

	return h.Process()
}

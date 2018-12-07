package main

import (
  "fmt"
  "context"
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-lambda-go/events"
  "github.com/aws/aws-lambda-go/lambda"
  "github.com/aws/aws-sdk-go/service/dynamodb"
  "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
  "github.com/aws/aws-sdk-go/service/s3"
  "github.com/aws/aws-sdk-go/service/s3/s3manager"
  "github.com/aws/aws-sdk-go/aws/session"
  "os"
  "encoding/csv"
)


var sess = session.Must(session.NewSessionWithOptions(session.Options{
  Config:            aws.Config{Region: aws.String(os.Getenv("AWS_REGION"))},
  SharedConfigState: session.SharedConfigEnable,
  Profile:           os.Getenv("AWS_PROFILE"),
}))

func handler(ctx context.Context, s3Event events.S3Event) {
  bucketFilePath := s3Event.Records[0].S3.Object.Key
  bucketName := s3Event.Records[0].S3.Bucket.Name

  filePath, err := DownloadCSVFromS3(bucketName, bucketFilePath)
  if err != nil {
    fmt.Errorf("%q, %v", filePath, err)
  }
  err = PopulateToDynamoDB(filePath)
  if err != nil {
    fmt.Printf("failed to read file, %v", err)
    return
  } 
  fmt.Printf("CSV %v successfully dumped to Dynamo", filePath)
}

func main() {
  lambda.Start(handler)
}

func PopulateToDynamoDB(filePath string) error {
  csvfile, err := os.Open(filePath)
  if err != nil {
    return fmt.Errorf("failed to open file, %v", err)
  }

  defer csvfile.Close()

  reader := csv.NewReader(csvfile)
  reader.FieldsPerRecord = -1 // see the Reader struct information below

  rawCSVdata, err := reader.ReadAll()

  if err != nil {
    fmt.Printf("failed to read file, %v", err)
  }

  if len(rawCSVdata) == 0 {
    fmt.Printf("Empty file, %v", filePath)
  }

  keys := rawCSVdata[0]

  for i, workerData := range rawCSVdata {
    workerDetail := make(map[string]string)
    // workerDetail.Id = generateUID()
    if i == 0 {
      // skip header line
      continue
    }

    for j, k := range keys {
      workerDetail[k] = workerData[j]
    }

    CreateRecord(workerDetail)
  }
  return nil
}

func CreateRecord(data map[string]string) {
  svc := dynamodb.New(sess)
  av, _ := dynamodbattribute.MarshalMap(data)

  input := &dynamodb.PutItemInput{
    Item:      av,
    TableName: aws.String(os.Getenv("TABLE_NAME")),
  }
  svc.PutItem(input)
  return
}

func DownloadCSVFromS3(bucket, bucketfilePath string) (string, error) {
  downloader := s3manager.NewDownloader(sess)
  filePath := fmt.Sprintf("/tmp/temp_file.csv")

  // Create a new local CSV file 
  csvfile, err := os.Create(filePath)
  if err != nil {
    return "", fmt.Errorf("failed to create file")
  }

  // Write the contents of S3 Object to the file
  _, err = downloader.Download(csvfile, &s3.GetObjectInput{
    Bucket: aws.String(bucket),
    Key:    aws.String(bucketfilePath),
  })

  if err != nil {
    return "", fmt.Errorf("Failed to download the file")
  }

  defer csvfile.Close()

  return filePath, nil
}

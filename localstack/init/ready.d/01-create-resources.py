
import boto3

ENDPOINT = "http://localhost:4566"
REGION = "us-east-1"

dynamodb = boto3.client(
    "dynamodb",
    endpoint_url=ENDPOINT,
    region_name=REGION,
    aws_access_key_id="test",
    aws_secret_access_key="test",
)

s3 = boto3.client(
    "s3",
    endpoint_url=ENDPOINT,
    region_name=REGION,
    aws_access_key_id="test",
    aws_secret_access_key="test",
)

# Create DynamoDB table
if "tasks" not in dynamodb.list_tables()["TableNames"]:
    dynamodb.create_table(
        TableName="tasks",
        KeySchema=[
            {"AttributeName": "id", "KeyType": "HASH"}
        ],
        AttributeDefinitions=[
            {"AttributeName": "id", "AttributeType": "N"}
        ],
        BillingMode="PAY_PER_REQUEST",
    )
    print("Created DynamoDB table: tasks")

# Create S3 bucket
buckets = [
    bucket["Name"]
    for bucket in s3.list_buckets()["Buckets"]
]

if "task-objects" not in buckets:
    s3.create_bucket(Bucket="task-objects")
    print("Created S3 bucket: task-objects")

print("AWS resources initialized successfully!")

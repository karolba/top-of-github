import os
from datetime import datetime, timedelta
from urllib.request import Request, urlopen
import boto3

MAX_DATA_AGE_DAYS = 7

SENDER_EMAIL_ADDRESS = os.environ['SENDER_EMAIL_ADDRESS']
TARGET_EMAIL_ADDRESS = os.environ['TARGET_EMAIL_ADDRESS']
METADATA_URL = os.environ['METADATA_URL']

def lambda_handler(event, context):
    with urlopen(Request(METADATA_URL, method='HEAD')) as response:
        last_modified_str = response.headers.get('Last-Modified')

    last_modified_date = datetime.strptime(last_modified_str, '%a, %d %b %Y %H:%M:%S %Z')
    current_date = datetime.utcnow()
    days_old = (current_date - last_modified_date).days

    if days_old > MAX_DATA_AGE_DAYS:
        email_content = f'The git-top-repos data fetched from Github has been last updated on {last_modified_str}.'
        send_email(f'git-top-repos data is {days_old} days old', email_content)

    return { 'statusCode': 200, 'body': '' }

def send_email(subject, message):
    ses = boto3.client('ses')

    response = ses.send_email(
        Destination={ 'ToAddresses': [TARGET_EMAIL_ADDRESS] },
        Message={
            'Body': {
                'Text': {
                    'Charset': 'UTF-8',
                    'Data': message,
                },
            },
            'Subject': {
                'Charset': 'UTF-8',
                'Data': subject,
            },
        },
        Source=SENDER_EMAIL_ADDRESS,
    )


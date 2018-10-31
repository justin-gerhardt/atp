#!/bin/ash

for file in ./*.mp3; do
    if [ -e "$file" ] ; then  # Make sure it isn't an empty match
        display=`basename "$file"`
        echo "Copying $display to S3 at ${OUTPUT_S3_PATH}/$display"
        aws s3 cp "$file" s3://${OUTPUT_S3_PATH} --region ${AWS_REGION}
    fi
done


# TLDR   

go run main.go


# DLQ Message Processor

This tool is designed for manual processing of Google Cloud Pub/Sub Dead Letter Queues (DLQs). It provides a command-line interface to handle messages that cannot be processed successfully after a certain number of attempts.

## Overview

Dead Letter Queues in Pub/Sub capture messages that fail to process after multiple attempts. This tool offers several modes of operation to manually manage these messages.

## Key Features

1. **Interactive Configuration**:
   - Users are prompted to input configuration details such as the source subscription path, processing mode, destination topic, and message limit. This ensures tailored processing to meet specific requirements.

2. **Processing Modes**:
   - **Preview Messages**: View messages in the DLQ without acknowledging them. Useful for inspecting message content and attributes.
   - **Republish to Another Topic**: Redirect messages to a specified destination topic for further processing or retry.
   - **Acknowledge Without Republishing**: Acknowledge and remove messages from the queue without further action.
   - **Custom Process Messages**: Implement custom logic to process messages, such as transforming data or applying business logic.

3. **Custom Processing Logic**:
   - Define custom processing logic in the `processCustomMessage` function. For example, modify the message path by replacing a domain before republishing.

4. **Resource Management**:
   - Manages Pub/Sub clients and topics, ensuring resources are properly initialized and closed. Supports handling different projects for source and destination topics.

5. **Concurrency Control**:
   - Processes messages sequentially by setting the maximum number of outstanding messages to one, ensuring individual handling for manual processing and debugging.

6. **Error Handling and Logging**:
   - Provides feedback on processing status through error handling and logging, helping users identify issues and understand outcomes.

## Use Cases

- **Debugging and Analysis**: Preview messages to diagnose issues and understand why they ended up in the DLQ.
- **Manual Intervention**: Manually republish or acknowledge messages based on analysis, handling exceptions that automated systems cannot resolve.
- **Custom Processing**: Apply specific business logic to messages for flexible handling of complex scenarios.

## Testing Custom Processing Logic

After making changes to the custom processing logic, you can adapt and  use the provided unit tests to ensure the desired outcome. The `processor_test.go` file includes tests for the `ProcessMessage` function. To run the tests, use the following command:

```bash
go test -v
```

This will execute the tests and verify that the message processing logic behaves as expected, including transformations and attribute preservation.

## Conclusion

This tool provides a comprehensive solution for manually managing Pub/Sub DLQs, offering flexibility and control over message processing. By using the interactive configuration and testing capabilities, users can effectively handle and process messages according to their specific needs.

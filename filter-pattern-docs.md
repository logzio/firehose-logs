<details>
  <summary>
    <h4>Filter Pattern Syntax</h4>
  </summary>

#### Using Filter Patterns with CloudWatch Logs

The `filterPattern` parameter allows you to specify which log events to send to Logz.io based on their content. Only logs that match the pattern will be forwarded. If left empty, all logs will be sent.

##### Basic Examples:

1. **Filter logs containing a specific term:**
   ```
   "ERROR"
   ```
   Matches log events that contain the word "ERROR"

2. **Filter logs containing multiple terms (AND condition):**
   ```
   "ERROR Exception"
   ```
   Matches log events that contain both "ERROR" and "Exception"

3. **Exclude logs with specific terms:**
   ```
   "ERROR" -"DEBUG"
   ```
   Matches log events that contain "ERROR" but not "DEBUG"

4. **JSON field filtering:**
   ```
   { $.level = "ERROR" }
   ```
   Matches JSON logs where the "level" field equals "ERROR"

##### Validation:

The system automatically validates the filter pattern syntax when you deploy the stack. If you provide an invalid pattern, the deployment will fail with a detailed error message.

For the complete syntax documentation, see [CloudWatch Logs Filter and Pattern Syntax](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html).

</details>

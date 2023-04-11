# Security 

There are a number of security concerns to keep in mind with instance data:

## Input & Output Data Is Remembered

A copy of the starting value is saved separately so that instances can be replayed. It also helps to debug workflows. It is worth keeping this in mind. Even if your workflow deletes sensitive fields, they are still stored somewhere they could be reused or read later. 

Likewise, output data is saved. It is returned to parent instances if the instance was executed as a subflow. Both input and output data is requestable via API. For a good way to handle most sensitive data, consider using secrets.

## Input Validation

To protect your workflows from behaving in unexpected ways, including intentional exploits by attackers, it is good practice to [validate](../workflow-yaml/validate.md) your input data before acting upon it. That is why we recommend beginning every workflow with a validate state.
 
## .private

As a basic precaution, anything stored under `.private` is redacted over the APIs that retrieve instance input and output data. This data is still usable by the instance. Could still be returned to a parent instance. It is still stored in the database in plaintext. And there is nothing preventing you from transforming it or passing it somewhere that exposes this information. Use this feature with caution. 
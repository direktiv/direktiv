import { RJSFSchema } from "@rjsf/utils";

export const basicExample: RJSFSchema = {
  title: "A registration form",
  type: "object",
  required: ["firstName", "lastName"],
  properties: {
    password: {
      type: "string",
      title: "Password",
    },
    lastName: {
      type: "string",
      title: "Last name",
    },
    bio: {
      type: "string",
      title: "Bio",
    },
    firstName: {
      type: "string",
      title: "First name",
    },
    age: {
      type: "integer",
      title: "Age",
    },
    occupation: {
      type: "string",
      enum: ["foo", "bar", "fuzz", "qux"],
    },
  },
};

export const arraySchemaSample: RJSFSchema = {
  definitions: {
    Thing: {
      type: "object",
      properties: {
        name: {
          type: "string",
          default: "Default name",
        },
      },
    },
  },
  type: "object",
  properties: {
    listOfStrings: {
      type: "array",
      title: "A list of strings",
      items: {
        type: "string",
        default: "bazinga",
      },
    },
    ChoicesList: {
      type: "array",
      title: "A choices list",
      description: "select one",
      items: {
        type: "string",
        enum: ["foo", "bar", "fuzz", "qux"],
      },
      uniqueItems: false,
    },
  },
};

export const customArraySample: RJSFSchema = {
  title: "Custom array of strings",
  type: "array",
  items: {
    type: "string",
  },
};
export const simpleSample: RJSFSchema = {
  title: "A registration form",
  description: "A simple form example description.",
  type: "object",
  required: ["firstName", "lastName"],
  properties: {
    firstName: {
      type: "string",
      title: "First name",
      description: "field description for first name",
    },
    lastName: {
      type: "string",
      title: "Last name",
    },
    age: {
      type: "integer",
      title: "Age",
    },
    bio: {
      type: "string",
      title: "Bio",
      format: "data-url",
    },
    password: {
      type: "string",
      title: "Password",
      minLength: 3,
    },
    telephone: {
      type: "string",
      title: "Telephone",
      minLength: 10,
    },
  },
};

export const exampleThatThrowsAnError: RJSFSchema = {
  title: "Selecting anything from the select field will show an error",
  type: "object",
  properties: {
    firstName: {
      type: "string",
      title: "First name",
    },
    select: {
      type: "integer",
      enum: ["foo", "bar", "fuzz", "qux"],
    },
  },
};

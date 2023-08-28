// takes a json input string and format it with 4 spaces indentation
export const prettifyJsonString = (jsonString: string) => {
  try {
    const resultAsJson = JSON.parse(jsonString);
    return JSON.stringify(resultAsJson, null, 4);
  } catch (e) {
    return "{}";
  }
};

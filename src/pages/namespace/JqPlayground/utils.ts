// takes a json input string and formats it as
export const prettifyJsonString = (jsonString: string) => {
  try {
    const resultAsJson = JSON.parse(jsonString);
    return JSON.stringify(resultAsJson, null, 4);
  } catch (e) {
    return "{}";
  }
};

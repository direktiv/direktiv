// TODO: add tests
export const parseDataUrl = (dataUrl: string) => {
  const splitUrl = dataUrl.split(";");
  if (!splitUrl || !splitUrl[0] || !splitUrl[1]) return null;

  const mimeType = splitUrl[0].split(":")[1];
  const data = splitUrl[1].split(",")[1];

  if (!mimeType || !data) return null;

  return {
    mimeType,
    data,
  };
};

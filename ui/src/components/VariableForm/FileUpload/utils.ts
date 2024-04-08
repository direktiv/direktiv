export const parseDataUrl = (dataUrl: string) => {
  const splitUrl = dataUrl.split(";");
  if (!splitUrl || !splitUrl[0] || !splitUrl[1]) return null;

  const mimeType = splitUrl[0].split(":")[1];
  const base64String = splitUrl[1].split(",")[1];

  if (!mimeType || !base64String) return null;

  return {
    mimeType,
    base64String,
  };
};

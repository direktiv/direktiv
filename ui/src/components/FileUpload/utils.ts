/**
 * takes a data url like
 *
 * data:application/json;base64,ewogICAgImNvb2xWaWRlbyI6ICJodHRwczovL3lvdXR1LmJlL29IZzVTSllSSEEwP3NpPXlRLVFLMEE1RlBiSG5rZDgiCn0=
 *
 * and returns an object with mimeType and base64String
 *
 * {
 *   base64String: "ewogICAgImNvb2xWaWRlbyI6ICJodHRwczovL3lvdXR1LmJlL29IZzVTSllSSEEwP3NpPXlRLVFLMEE1RlBiSG5rZDgiCn0=",
 *   mimeType: "application/json",
 * }
 *
 */
export const parseDataUrl = (dataUrl: string) => {
  const dataUrlArr = dataUrl.split(";");
  if (!dataUrlArr[0] || !dataUrlArr[1]) return null;

  const mimeType = dataUrlArr[0].split(":")[1];
  const base64String = dataUrlArr[1].split(",")[1];

  if (!mimeType || !base64String) return null;

  return {
    mimeType,
    base64String,
  };
};

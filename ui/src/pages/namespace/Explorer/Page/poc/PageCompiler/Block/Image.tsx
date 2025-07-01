import { Card } from "~/design/Card";
import { ImageType } from "../../schema/blocks/image";
import { useStringInterpolation } from "../primitives/Variable/utils/useStringInterpolation";

type ImageProps = {
  blockProps: ImageType;
};

export const Image = ({ blockProps }: ImageProps) => {
  const { src, width, height } = blockProps;
  const interpolateString = useStringInterpolation();
  const interPolatedSrc = interpolateString(src);
  return <img src={interPolatedSrc} width={width} height={height} />;
};

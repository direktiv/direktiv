import { AllBlocksType } from "../../schema/blocks";
import { Headline } from "./Headline";
import { Text } from "./Text";
import { TwoColumns } from "./TwoColumns";

type BlockProps = {
  block: AllBlocksType;
};

export const Block = ({ block }: BlockProps) => {
  switch (block.type) {
    case "headline":
      return <Headline {...block} />;
      break;
    case "text":
      return <Text {...block} />;
      break;
    case "two-columns":
      return <TwoColumns {...block} />;
      break;
    default:
      return <div>not implemented yet: {block.type}</div>;
  }
};

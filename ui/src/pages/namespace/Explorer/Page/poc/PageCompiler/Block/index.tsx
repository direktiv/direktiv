import { AllBlocksType } from "../../schema/blocks";
import { BlockWrapper } from "./utils/BlockWrapper";
import { Headline } from "./Headline";
import { QueryProvider } from "./QueryProvider";
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
    case "query-provider":
      return <QueryProvider {...block} />;
      break;

    default:
      return <BlockWrapper>not implemented yet: {block.type}</BlockWrapper>;
  }
};

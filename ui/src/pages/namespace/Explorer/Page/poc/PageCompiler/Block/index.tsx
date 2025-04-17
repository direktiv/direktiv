import { AllBlocksType } from "../../schema/blocks";

type BlockProps = {
  block: AllBlocksType;
};

export const Block = ({ block }: BlockProps) => {
  switch (block.type) {
    case "headline":
      return <div className="text-xl">headline</div>;
      break;
    case "two-columns":
      return (
        <div className="grid grid-cols-2">
          <div>
            {block.leftBlocks.map((block, index) => (
              <Block key={index} block={block} />
            ))}
          </div>
          <div>
            {block.rightBlocks.map((block, index) => (
              <Block key={index} block={block} />
            ))}
          </div>
        </div>
      );
      break;
    default:
      return <div>not implemented yet: {block.type}</div>;
  }
};

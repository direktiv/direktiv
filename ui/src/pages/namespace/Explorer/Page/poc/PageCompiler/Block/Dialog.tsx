import { Block, BlockPathType } from ".";
import {
  DialogContent,
  Dialog as DialogDesignComponent,
  DialogTrigger,
} from "~/design/Dialog";

import { BlockList } from "./utils/BlockList";
import { Button } from "./Button";
import { DialogType } from "../../schema/blocks/dialog";

type DialogProps = {
  blockProps: DialogType;
  blockPath: BlockPathType;
};
export const Dialog = ({ blockProps, blockPath }: DialogProps) => {
  const { blocks, trigger } = blockProps;
  return (
    <DialogDesignComponent>
      <DialogTrigger asChild>
        <Button blockProps={trigger} />
      </DialogTrigger>
      <DialogContent showCloseButton>
        <BlockList>
          {blocks.map((block, index) => (
            <Block
              key={index}
              block={block}
              blockPath={[...blockPath, index]}
            />
          ))}
        </BlockList>
      </DialogContent>
    </DialogDesignComponent>
  );
};

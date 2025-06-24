import { Block, BlockPathType } from ".";
import {
  DialogClose,
  DialogContent,
  Dialog as DialogDesignComponent,
  DialogFooter,
  DialogTrigger,
} from "~/design/Dialog";

import { BlockList } from "./utils/BlockList";
import { Button } from "./Button";
import ButtonDesignComponent from "~/design/Button";
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
      <DialogContent>
        <BlockList path={blockPath}>
          {blocks.map((block, index) => (
            <Block
              key={index}
              block={block}
              blockPath={[...blockPath, index]}
            />
          ))}
        </BlockList>
        <DialogFooter>
          <DialogClose asChild>
            <ButtonDesignComponent variant="ghost">
              Cancel
            </ButtonDesignComponent>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </DialogDesignComponent>
  );
};

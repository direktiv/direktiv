import { Block, BlockPathType } from "..";
import { LocalDialog, LocalDialogContent } from "~/design/LocalDialog";
import { RowActionsType, TableActionsType } from "../../../schema/blocks/table";

import { BlockList } from "../utils/BlockList";
import { DialogXClose } from "~/design/Dialog";
import { useState } from "react";

type ActionsDialogProps = {
  actions: RowActionsType | TableActionsType;
  blockPath: BlockPathType;

  renderTrigger: (
    setOpenedDialogIndex: (index: number | null) => void
  ) => React.ReactNode;
};

export const ActionsDialog = ({
  actions,
  blockPath,

  renderTrigger,
}: ActionsDialogProps) => {
  const [openedDialogIndex, setOpenedDialogIndex] = useState<number | null>(
    null
  );

  return (
    <LocalDialog open={openedDialogIndex !== null}>
      {renderTrigger(setOpenedDialogIndex)}
      <div onClick={(event) => event.preventDefault()}>
        <LocalDialogContent>
          <DialogXClose onClick={() => setOpenedDialogIndex(null)} />
          <div className="max-h-[55vh] overflow-y-auto p-2 pt-4">
            {openedDialogIndex !== null && (
              <BlockList path={[...blockPath, openedDialogIndex]}>
                {(actions.blocks[openedDialogIndex]?.blocks ?? []).map(
                  (block, idx) => (
                    <Block
                      key={idx}
                      block={block}
                      blockPath={[...blockPath, openedDialogIndex, idx]}
                    />
                  )
                )}
              </BlockList>
            )}
          </div>
        </LocalDialogContent>
      </div>
    </LocalDialog>
  );
};

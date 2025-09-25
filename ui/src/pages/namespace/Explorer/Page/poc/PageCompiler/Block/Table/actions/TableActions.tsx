import { ActionsDialog } from "./ActionsDialog";
import { BlockPathType } from "../..";
import ButtonDesignComponent from "~/design/Button";
import { TableActionsType } from "../../../../schema/blocks/table";

type TableActionsProps = {
  actions: TableActionsType;
  blockPath: BlockPathType;
};

export const TableActions = ({ actions, blockPath }: TableActionsProps) => (
  <ActionsDialog
    actions={actions}
    blockPath={[...blockPath, 0]}
    renderTrigger={(setOpenedDialogIndex) => (
      <div className="flex gap-2">
        {actions.blocks.map((dialog, index) => (
          <ButtonDesignComponent
            key={index}
            size="sm"
            onClick={(event) => {
              event.stopPropagation();
              setOpenedDialogIndex(index);
            }}
          >
            {dialog.trigger.label}
          </ButtonDesignComponent>
        ))}
      </div>
    )}
  />
);

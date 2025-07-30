import { Code, Eye, LucideIcon, Pencil } from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import { ButtonBar } from "~/design/ButtonBar";
import { FC } from "react";
import { PageEditorMode } from ".";
import { Toggle } from "~/design/Toggle";
import { useTranslation } from "react-i18next";

type Button = {
  id: PageEditorMode;
  icon: LucideIcon;
};

const buttons: Button[] = [
  {
    id: "code",
    icon: Code,
  },
  {
    id: "edit",
    icon: Pencil,
  },
  {
    id: "live",
    icon: Eye,
  },
];

type EditorModeSwitcherProps = {
  value: PageEditorMode;
  onChange: (value: PageEditorMode) => void;
};

const EditorModeSwitcher: FC<EditorModeSwitcherProps> = ({
  value,
  onChange,
}) => {
  const { t } = useTranslation();
  return (
    <ButtonBar>
      <TooltipProvider>
        {buttons.map((button) => {
          const IconComponent = button.icon;
          return (
            <Tooltip key={button.id}>
              <TooltipTrigger asChild>
                <div className="flex grow">
                  <Toggle
                    onClick={() => {
                      onChange(button.id);
                    }}
                    className="grow"
                    pressed={button.id === value}
                  >
                    <IconComponent />
                  </Toggle>
                </div>
              </TooltipTrigger>
              <TooltipContent>
                {t(`direktivPage.blockEditor.toolbar.${button.id}`)}
              </TooltipContent>
            </Tooltip>
          );
        })}
      </TooltipProvider>
    </ButtonBar>
  );
};

export default EditorModeSwitcher;

import { Code, Eye, LucideIcon, Pencil } from "lucide-react";
import { FC, useState } from "react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import { ButtonBar as DesignButtonBar } from "~/design/ButtonBar";
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

type ButtonBarProps = {
  value: PageEditorMode;
  onChange: (value: PageEditorMode) => void;
};

/**
 *
 * TODO: rename comopnent and file
 * any better way to synch state
 *
 */
const ButtonBar: FC<ButtonBarProps> = ({ value, onChange }) => {
  const [activeButton, setActiveButton] = useState<PageEditorMode>(value);
  const { t } = useTranslation();
  return (
    <DesignButtonBar>
      <TooltipProvider>
        {buttons.map((button) => {
          const IconComponent = button.icon;
          return (
            <Tooltip key={button.id}>
              <TooltipTrigger asChild>
                <div className="flex grow">
                  <Toggle
                    onClick={() => {
                      setActiveButton(button.id);
                      onChange(button.id);
                    }}
                    className="grow"
                    pressed={button.id === activeButton}
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
    </DesignButtonBar>
  );
};

export default ButtonBar;

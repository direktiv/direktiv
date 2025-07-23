import { Code, Eye, LucideIcon, Pencil } from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import { ButtonBar as DesignButtonBar } from "~/design/ButtonBar";
import { PageCompilerMode } from "./poc/PageCompiler/context/pageCompilerContext";
import { Toggle } from "~/design/Toggle";
import { useState } from "react";
import { useTranslation } from "react-i18next";

type ButtonState = PageCompilerMode | "code";

type Button = {
  id: ButtonState;
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

const ButtonBar = () => {
  const [activeButton, setActiveButton] = useState<ButtonState>("edit");
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
                    onClick={() => setActiveButton(button.id)}
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

import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";
import {
  availableLayouts,
  layoutIcons,
  useEditorActions,
  useEditorLayout,
} from "~/util/store/editor";

import { ButtonBar } from "~/design/ButtonBar";
import { Toggle } from "~/design/Toggle";
import { useTranslation } from "react-i18next";

export const EditorLayoutSwitcher = () => {
  const { t } = useTranslation();
  const currentLayout = useEditorLayout();
  const { setLayout: setCurrentLayout } = useEditorActions();
  return (
    <ButtonBar>
      <TooltipProvider>
        {availableLayouts.map((layout) => {
          const Icon = layoutIcons[layout];
          return (
            <Tooltip key={layout}>
              <TooltipTrigger asChild>
                <div className="flex grow">
                  <Toggle
                    onClick={() => {
                      setCurrentLayout(layout);
                    }}
                    className="grow"
                    pressed={layout === currentLayout}
                    data-testid={`editor-layout-btn-${layout}`}
                  >
                    <Icon />
                  </Toggle>
                </div>
              </TooltipTrigger>
              <TooltipContent>
                {t(`pages.explorer.workflow.editor.layout.${layout}`)}
              </TooltipContent>
            </Tooltip>
          );
        })}
      </TooltipProvider>
    </ButtonBar>
  );
};

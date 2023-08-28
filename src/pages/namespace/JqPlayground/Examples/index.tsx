import { Bookmark, Play } from "lucide-react";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import exampleSnippets from "./exampleSnippets";
import { useTranslation } from "react-i18next";

const Examples = ({
  onExampleClick,
}: {
  onExampleClick: (params: { query: string; input: string }) => void;
}) => {
  const { t } = useTranslation();
  return (
    <Card className="flex flex-col gap-5 p-5">
      <h3 className="flex grow items-center gap-x-2 font-medium">
        <Bookmark className="h-5" />
        {t("pages.jqPlayground.examples.title")}
      </h3>
      <div className="grid grid-cols-2 gap-5 text-sm">
        {exampleSnippets.map(({ query, key, input, example }) => (
          <Card key={key} className="flex gap-2 p-5">
            <div className="grid grow grid-cols-2">
              <div className="font-mono text-primary-500">{example}</div>
              <div>{t(`pages.jqPlayground.examples.snippets.${key}`)}</div>
            </div>
            <Button
              size="sm"
              variant="outline"
              onClick={() =>
                onExampleClick({
                  query,
                  input,
                })
              }
            >
              <Play />
              {t("pages.jqPlayground.examples.buttonLabel")}
            </Button>
          </Card>
        ))}
      </div>
    </Card>
  );
};

export default Examples;

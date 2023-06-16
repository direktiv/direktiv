import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { z } from "zod";

const mimeTypes = [
  { label: "JSON", value: "application/json" },
  { label: "YAML", value: "application/yaml" },
  { label: "shell", value: "application/x-sh" },
  { label: "plaintext", value: "text/plain" },
  { label: "HTML", value: "text/html" },
  { label: "CSS", value: "text/css" },
];

export const mimeTypeToLanguageDict = {
  "application/json": "json",
  "application/yaml": "yaml",
  "application/x-sh": "shell",
  "text/plain": "plaintext",
  "text/html": "html",
  "text/css": "css",
} as const;

export const MimeTypeSchema = z.enum([
  "application/json",
  "application/yaml",
  "application/x-sh",
  "text/plain",
  "text/html",
  "text/css",
]);

export type MimeTypeType = z.infer<typeof MimeTypeSchema>;

const MimeTypeSelect = ({
  mimeType,
  onChange,
}: {
  mimeType: string | undefined;
  onChange: (value: MimeTypeType) => void;
}) => (
  <Select value={mimeType} onValueChange={onChange}>
    <SelectTrigger variant="outline">
      <SelectValue placeholder="Select a mimetype" />
    </SelectTrigger>
    <SelectContent>
      {mimeTypes.map((type) => (
        <SelectItem key={type.value} value={type.value}>
          {type.label}
        </SelectItem>
      ))}
    </SelectContent>
  </Select>
);

export default MimeTypeSelect;

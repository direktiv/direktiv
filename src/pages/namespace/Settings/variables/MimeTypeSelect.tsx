import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

const mimeTypes = [
  { label: "JSON", value: "application/json" },
  { label: "YAML", value: "application/yaml" },
  { label: "shell", value: "application/x-sh" },
  { label: "plaintext", value: "text/plain" },
  { label: "HTML", value: "text/html" },
  { label: "CSS", value: "text/css" },
];

const MimeTypeSelect = ({
  mimeType,
  onChange,
}: {
  mimeType: string | undefined;
  onChange: (value: string) => void;
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

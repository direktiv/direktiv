import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../Select";

import Button from "../Button";

export default {
  title: "Components (next)/Compositions/Form",
};

export const Default = () => (
  <div className="card bg-base-100 p-6 shadow-md">
    <div className="mt-6 grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
      <div className="form-control sm:col-span-4">
        <label className="label">
          <span className="label-text">Some text input</span>
        </label>
        <input
          type="text"
          placeholder="text"
          className="input-bordered input w-full"
        />
      </div>
      <div className="form-control sm:col-span-2">
        <label className="label">
          <span className="label-text">Another text input</span>
          <span className="label-text-alt">required</span>
        </label>
        <input
          type="text"
          placeholder="text"
          className="input-bordered input w-full"
        />
      </div>
      <div className="form-control sm:col-span-4">
        <div className="form-control sm:col-span-4">
          <label className="label">
            <span className="label-text">Some text input</span>
          </label>
          <input
            type="text"
            placeholder="text"
            className="input-bordered input w-full"
          />
        </div>
      </div>
      <div className="form-control sm:col-span-2">
        <label className="label">
          <span className="label-text">Select something</span>
        </label>
        <Select>
          <SelectTrigger>
            <SelectValue placeholder="block element" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="1">Item 1</SelectItem>
            <SelectItem value="2">Item 2</SelectItem>
            <SelectItem value="3">Item 3</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <div className="flex flex-row-reverse gap-5 sm:col-span-full">
        <Button variant="primary">Submit</Button>
        <Button variant="ghost">Cancel</Button>
      </div>
    </div>
  </div>
);

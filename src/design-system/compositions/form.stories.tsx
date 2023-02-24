import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../select";
import Button from "../button";

export default {
  title: "Design System/Compositions/Form",
  parameters: {
    docs: {
      page: null,
    },
  },
};

export const Default = () => (
  <div className="card bg-base-100 shadow-md p-6">
    <div className="mt-6 grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
      <div className="sm:col-span-4 form-control">
        <label className="label">
          <span className="label-text">Some text input</span>
        </label>
        <input
          type="text"
          placeholder="text"
          className="input input-bordered w-full"
        />
      </div>
      <div className="sm:col-span-2 form-control">
        <label className="label">
          <span className="label-text">Another text input</span>
          <span className="label-text-alt">required</span>
        </label>
        <input
          type="text"
          placeholder="text"
          className="input input-bordered w-full"
        />
      </div>
      <div className="sm:col-span-4 form-control">
        <div className="sm:col-span-4 form-control">
          <label className="label">
            <span className="label-text">Some text input</span>
          </label>
          <input
            type="text"
            placeholder="text"
            className="input input-bordered w-full"
          />
        </div>
      </div>
      <div className="sm:col-span-2 form-control">
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
      <div className="sm:col-span-full flex flex-row-reverse gap-5">
        <Button color="primary">Submit</Button>
        <Button color="ghost">Cancel</Button>
      </div>
    </div>
  </div>
);

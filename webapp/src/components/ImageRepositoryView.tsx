import React from "react";
import { Divider } from "primereact/divider";
import { Dropdown } from "primereact/dropdown";
import { DataScroller } from "primereact/datascroller";
import ImageTagView from "./ImageTagView";
import ImageTagViewCard from "./ImageTagViewCard";
import { Button } from "primereact/button";
import { InputText } from "primereact/inputtext";

type ImageRepositoryViewProps = {
  registery: string;
  namespace: string;
  image_repository: string;
};

const ImageRepositoryView = (props: ImageRepositoryViewProps) => {
  return (
    <div className="h-full flex flex-column pl-2 pr-2  w-full">
      <div className="flex justify-content-between pl-4 pr-4">
        <div className="flex align-items-center gap-2 text-sm">
          <div>Sort By</div>
          <div>
            <Dropdown
              options={[
                { name: "Recently updated" },
                { name: "Most pulled" },
                { name: "Image Size" },
              ]}
              optionLabel="name"
              className="w-full md:w-14rem border-1"
            />
          </div>
          <div>

          </div>
        </div>
        <div className="flex align-items-center gap-5">
          <div className="flex flex-column gap-2">
            <div className="pi pi-tags text-lg"></div>
            <div>2</div>
          </div>
          <div className="flex flex-column gap-2">
            <div className="pi pi-server text-lg"></div>
            <div>2.3 Gb</div>
          </div>
        </div>
      </div>
      <Divider className="pt-0 mt-3 w-full" />

      {/* <DataScroller value={[]} itemTemplate={(item:any) => ImageTagViewCard}/> */}
    </div>
  );
};

export default ImageRepositoryView;

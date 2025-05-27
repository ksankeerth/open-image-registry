import { Divider } from "primereact/divider";
import React from "react";

type ImageCardViewProps = {
  data: {
    upstreamDomain: string;
    registeryName: string;
    isCachedImage: boolean;
    namespace: string;
    imageRepository: string;
    cachedTags: string[];
    pulls: number;
    mostPulledTag: string;
    leastPulledTag: string;
    storageSpaceOfAllTags: number; // in GBs
    lastUpdatedAt: Date;
  };
  controls: {
    showRegistery: boolean;
  };
};

const ImageCardView = (props: ImageCardViewProps) => {
  return (
    <div className="border-teal-100 border-solid border-1 border-round-md shadow-1 pl-3 pr-3 pb-2">
      <div className=" flex justify-content-between pt-3  gap-2">
        <div className="text-color font-medium">
          {(props.data.namespace + "/" + props.data.imageRepository).substring(
            0,
            30
          )}
        </div>
        <div>
          {props.data.isCachedImage && (
            <i className="pi pi-sync" style={{ fontSize: "0.8rem" }}></i>
          )}
          {!props.data.isCachedImage && (
            <i className="pi pi-server" style={{ fontSize: "0.8rem" }}></i>
          )}
        </div>
      </div>
      <div className="grid grid-nogutter pt-3 ">
        <div className="col-12">
          <div className="bg-json-or-code-view p-2 border-round-md">
            {props.data.upstreamDomain}
          </div>
        </div>

        <div className="col-3 text-sm pt-3">Tags</div>
        <div
          className="col-9 pt-3 flex align-items-center gap-2"
          style={{ fontSize: "0.75rem" }}
        >
          <div className="flex align-items-center">
            <div className="underline text-blue-300">0.7&nbsp;</div>
            <div>
              (<i style={{ fontSize: "0.70rem" }} className="pi pi-download"></i> 11)
            </div>
          </div>

          <div className="text-blue-300">...........</div>
          <div className="flex align-items-center">
            <div className="underline text-blue-300">0.3&nbsp;</div>
            <div>
              (<i style={{ fontSize: "0.70rem" }} className="pi pi-download"></i> 3)
            </div>
          </div>
        </div>
        <Divider className="mb-0"/>
        <div className="col-6 pt-3 text-sm">
          Storage
        </div>
      
        <div className="col-6 pt-3 text-sm flex justify-content-end">
          Image Size(avg)
        </div>
        <div className="col-8 pt-2 text-sm ">
          1.8 GB
        </div>
        <div className="col-4 pt-2 text-sm flex justify-content-end">
          455 MB
        </div>

        {/* <div className="col-12 pt-3 flex justify-content-between">
          <div className="flex align-items-center cursor-pointer">
            <span className="text-sm">{props.data.pulls}</span>&nbsp;&nbsp;{" "}
            <i className="pi pi-download" />
          </div>
          <Divider layout="vertical" className="m-0 p-0" />
          <div className="flex align-items-center justify-content-end cursor-pointer">
            <span className="text-sm">
              {props.data.storageSpaceOfAllTags} GB
            </span>
            &nbsp;&nbsp; <i className="pi pi-gauge" />
          </div>
        </div> */}
      </div>
    </div>
  );
};

export default ImageCardView;

import React, { useLayoutEffect, useRef, useState } from "react";
import { Checkbox } from "primereact/checkbox";
import { MultiSelect } from "primereact/multiselect";
import { RadioButton } from "primereact/radiobutton";
import { ToggleButton } from "primereact/togglebutton";
import { Button } from "primereact/button";
import ImageCardView from "./ImageCardView";

const ImagesViewSummary = () => {
  const topElemRef = useRef<HTMLDivElement>(null);
  const bottomElemRef = useRef<HTMLDivElement>(null);

  const [contentHeight, setContentHeight] = useState<Number>(0);

  useLayoutEffect(() => {
    const calculateHeight = () => {
      const viewportHeight = window.innerHeight;
      const y1 =
        Number(topElemRef.current?.offsetTop) +
        Number(topElemRef.current?.offsetHeight);
      const y2 = Number(bottomElemRef.current?.offsetTop);
      const availableHeight = y2 - y1;
      setContentHeight(availableHeight);
    };
    calculateHeight();
    window.addEventListener("resize", calculateHeight);
    return () => window.removeEventListener("resize", calculateHeight);
  }, []);

  return (
    <div className="h-full flex flex-column justify-content-between pl-2 pr-2  w-full">
      {/* Filter */}
      <div className="w-full flex align-items-center justify-content-around gap-3 border-solid border-round-md surface-border p-2 text-sm">
        <div className="">
          <Checkbox
            inputId="images_filter_only_cached_images"
            name="images_filter_only_cached_images"
            value=""
            checked={true}
          />
          <label
            htmlFor="images_filter_only_cached_images"
            className="ml-2 text-sm"
          >
            Show only cached images
          </label>
        </div>
        <div>
          <MultiSelect
            options={[{ name: "proxy-quay-registry" }]}
            optionLabel="name"
            placeholder="Select Registeries"
            maxSelectedLabels={3}
            className="w-full md:w-20rem text-sm border-1"
            size={5}
          />
        </div>
        <div>Sort By:</div>
        <div className="flex flex-column gap-2">
          <div>
            <RadioButton
              inputId="last_pulled_at"
              name="sort_by"
              value="last_pulled_at"
            />
            <label htmlFor="last_pulled_at" className="ml-2">
              Recently pulled
            </label>
          </div>
          <div>
            <RadioButton
              inputId="last_updated_at"
              name="sort_by"
              value="last_updated_at"
              checked
              variant="outlined"
            />
            <label htmlFor="last_updated_at" className="ml-2">
              Recently updated
            </label>
          </div>
          <div>
            <RadioButton
              inputId="most_pulled"
              name="sort_by"
              value="most_pulled"
              variant="outlined"
            />
            <label htmlFor="most_pulled" className="ml-2">
              Most pulled
            </label>
          </div>
          <div>
            <RadioButton
              inputId="storage_space"
              name="sort_by"
              value="storage_space"
              variant="outlined"
            />
            <label htmlFor="storage_space" className="ml-2">
              Storage Space
            </label>
          </div>
        </div>
        <div>Order:</div>
        <div className="flex border-1 border-solid surface-border border-round-xl">
          <div className="pi pi-sort-amount-up p-2 border-1 border-solid surface-border border-round-left-xl"></div>

          <div className="pi pi-sort-amount-down p-2 border-1 border-solid surface-border border-round-right-xl"></div>
        </div>
      </div>

      {/* Content */}
      <div className="w-full h-full flex flex-row justify-content-between align-items-center gap-2">
        <div className="flex  align-items-center  hover:surface-50 border-round-md p-1 cursor-pointer">
          <div className="pi pi-angle-left text-xl surface-50  border-round-md p-2"></div>
        </div>
        <div
          className="pt-3 flex-grow-1 flex flex-column justify-content-around"
          style={{
            height: `${Number(contentHeight) - 50}px`,
            overflowY: "auto",
          }}
        >
          {/* Here we should only limited cards which can be seen in the screen and avoid adding scroll at x and y directions. */}
          <div className="flex gap-3">
            <ImageCardView
              data={{
                registeryName: "Quay",
                namespace: "tensorflow",
                imageRepository: "tensorflow",
                isCachedImage: true,
                cachedTags: ["0.3", "0.7"],
                pulls: 123,
                mostPulledTag: "0.7",
                leastPulledTag: "0.3",
                lastUpdatedAt: new Date(),
                storageSpaceOfAllTags: 1.8,
                upstreamDomain: "quay.io",
              }}
              controls={{ showRegistery: true }}
            />
            <ImageCardView
              data={{
                registeryName: "Quay",
                namespace: "tensorflow",
                imageRepository: "tensorflow",
                isCachedImage: true,
                cachedTags: ["0.3", "0.7"],
                pulls: 123,
                mostPulledTag: "0.7",
                leastPulledTag: "0.3",
                lastUpdatedAt: new Date(),
                storageSpaceOfAllTags: 1.8,
                upstreamDomain: "quay.io",
              }}
              controls={{ showRegistery: true }}
            />
            <ImageCardView
              data={{
                registeryName: "Quay",
                namespace: "tensorflow",
                imageRepository: "tensorflow",
                isCachedImage: true,
                cachedTags: ["0.3", "0.7"],
                pulls: 123,
                mostPulledTag: "0.7",
                leastPulledTag: "0.3",
                lastUpdatedAt: new Date(),
                storageSpaceOfAllTags: 1.8,
                upstreamDomain: "quay.io",
              }}
              controls={{ showRegistery: true }}
            />
          </div>
          <div className="flex gap-3 pt-4">
            <ImageCardView
              data={{
                registeryName: "Quay",
                namespace: "tensorflow",
                imageRepository: "tensorflow",
                isCachedImage: true,
                cachedTags: ["0.3", "0.7"],
                pulls: 123,
                mostPulledTag: "0.7",
                leastPulledTag: "0.3",
                lastUpdatedAt: new Date(),
                storageSpaceOfAllTags: 1.8,
                upstreamDomain: "quay.io",
              }}
              controls={{ showRegistery: true }}
            />
            <ImageCardView
              data={{
                registeryName: "Quay",
                namespace: "tensorflow",
                imageRepository: "tensorflow",
                isCachedImage: true,
                cachedTags: ["0.3", "0.7"],
                pulls: 123,
                mostPulledTag: "0.7",
                leastPulledTag: "0.3",
                lastUpdatedAt: new Date(),
                storageSpaceOfAllTags: 1.8,
                upstreamDomain: "quay.io",
              }}
              controls={{ showRegistery: true }}
            />
            <ImageCardView
              data={{
                registeryName: "Quay",
                namespace: "tensorflow",
                imageRepository: "tensorflow",
                isCachedImage: true,
                cachedTags: ["0.3", "0.7"],
                pulls: 123,
                mostPulledTag: "0.7",
                leastPulledTag: "0.3",
                lastUpdatedAt: new Date(),
                storageSpaceOfAllTags: 1.8,
                upstreamDomain: "quay.io",
              }}
              controls={{ showRegistery: true }}
            />
          </div>
        </div>
        <div className="flex  align-items-center  hover:surface-50 border-round-md p-1 cursor-pointer">
          <div className="pi pi-angle-right text-xl surface-50  border-round-md p-2"></div>
        </div>
      </div>
      <div ref={bottomElemRef}></div>
    </div>
  );
};

export default ImagesViewSummary;

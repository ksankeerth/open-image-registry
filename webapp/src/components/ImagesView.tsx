import React, { useEffect, useRef, useState } from "react";
import { Sidebar } from "primereact/sidebar";
import LogoComponent from "./LogoComponent";
import { Splitter, SplitterPanel } from "primereact/splitter";
import { TreeNode } from "primereact/treenode";
import ImageTagsTreeView from "./ImageTagsTreeView";
import { ImageTagsTreeNodeTypes } from "../types/app_types";
import { Button } from "primereact/button";
import { InputText } from "primereact/inputtext";
import ImagesViewSummary from "./ImagesViewSummary";
import ImageRepositoryView from "./ImageRepositoryView";

const ImagesView = (props: {
  visible: boolean;
  hideCallback: React.Dispatch<React.SetStateAction<boolean>>;
}) => {
  const [imageTagsTree, setImageTagsTree] = useState<TreeNode[]>([]);

  useEffect(() => {
    imageTagsTree.push(
      ...[
        {
          id: "proxy-docker-hub",
          key: "proxy-docker-hub",
          label: "proxy-docker-hub",
          data: {
            node_type: ImageTagsTreeNodeTypes.Registry,
            registery: "proxy-docker-hub",
            namespace: "",
            image_repository: "",
            tag: ""
          },
          children: [
            {
              id: "proxy-docker-hub/tensorflow",
              key: "proxy-docker-hub/tensorflow",
              label: "tensorflow",
              data: {
                node_type: ImageTagsTreeNodeTypes.Namespace,
                registery: "proxy-docker-hub",
                namespace: "tensorflow",
                image_repository: "",
                tag: ""
              },
              children: [
                {
                  id: "proxy-docker-hub/tensorflow/tensorflow",
                  key: "proxy-docker-hub/tensorflow/tensorflow",
                  label: "tensorflow",
                  data: {
                    node_type: ImageTagsTreeNodeTypes.Repository,
                    registery: "proxy-docker-hub",
                    namespace: "tensorflow",
                    image_repository: "tensorflow",
                    tag: ""
                  },
                  children: [
                    {
                      id: "proxy-docker-hub/tensorflow/tensorflow/0.1",
                      key: "proxy-docker-hub/tensorflow/tensorflow/0.1",
                      label: "0.1",
                      data: {
                        node_type: ImageTagsTreeNodeTypes.Tag,
                        registery: "proxy-docker-hub",
                        namespace: "tensorflow",
                        image_repository: "tensorflow",
                        tag: "0.1"
                      },
                      leaf: true,
                    },
                    {
                      id: "proxy-docker-hub/tensorflow/tensorflow/0.2",
                      key: "proxy-docker-hub/tensorflow/tensorflow/0.2",
                      label: "0.2",
                      data: {
                        node_type: ImageTagsTreeNodeTypes.Tag,
                        registery: "proxy-docker-hub",
                        namespace: "tensorflow",
                        image_repository: "tensorflow",
                        tag: "0.2"
                      },
                      leaf: true,
                    },
                    {
                      id: "proxy-docker-hub/tensorflow/tensorflow/0.3",
                      key: "proxy-docker-hub/tensorflow/tensorflow/0.3",
                      label: "0.3",
                      data: {
                        node_type: ImageTagsTreeNodeTypes.Tag,
                        registery: "proxy-docker-hub",
                        namespace: "tensorflow",
                        image_repository: "tensorflow",
                        tag: "0.3"
                      },
                      leaf: true,
                    },
                    {
                      id: "proxy-docker-hub/tensorflow/tensorflow/0.4",
                      key: "proxy-docker-hub/tensorflow/tensorflow/0.4",
                      label: "0.4",
                      data: {
                        node_type: ImageTagsTreeNodeTypes.Tag,
                      },
                      leaf: true,
                    },
                    {
                      id: "proxy-docker-hub/tensorflow/tensorflow/0.5",
                      key: "proxy-docker-hub/tensorflow/tensorflow/0.5",
                      label: "0.5",
                      data: {
                        node_type: ImageTagsTreeNodeTypes.Tag,
                        registery: "proxy-docker-hub",
                        namespace: "tensorflow",
                        image_repository: "tensorflow",
                        tag: "0.5"
                      },
                      leaf: true,
                    },
                    {
                      id: "proxy-docker-hub/tensorflow/tensorflow/0.6",
                      key: "proxy-docker-hub/tensorflow/tensorflow/0.6",
                      label: "0.6",
                      data: {
                        node_type: ImageTagsTreeNodeTypes.Tag,
                        registery: "proxy-docker-hub",
                        namespace: "tensorflow",
                        image_repository: "tensorflow",
                        tag: "0.6"
                      },
                      leaf: true,
                    },
                  ],
                },
              ],
            },
          ],
        },
      ]
    );
  });

  const showImageRepositoryView = (registery: string, namespace: string, imageRepository: string) => {

  }

  const showImageTagView = (registery: string, namespace: string, imageRepository: string, tag: string) => {

  }


  return (
    <Sidebar
      position="bottom"
      visible={props.visible}
      onHide={() => props.hideCallback(false)}
      header={
        <div className="flex flex-row w-full surface-0">
          <div className="flex flex-grow-1">
            <LogoComponent />
          </div>
          <div className="flex-grow-0 flex align-items-center justify-content-center font-semibold">
            {/* Title shoud change with view we shows in right spliter panel */}
            {/* Images  */}
            IMAGE: proxy-docker-hub/tensorflow/tensorflow
          </div>
          <div className="flex-grow-1" style={{ visibility: "hidden" }}>
            <LogoComponent />
          </div>
        </div>
      }
      style={{ height: "100vh", overflow: "hidden" }}
      className="flex align-items-stretch"
    >
      <div
        className="flex flex-column gap-2 h-full"
        style={{ overflow: "hidden" }}
      >
        <Splitter className="border-none flex-grow-1">
          <SplitterPanel size={5} minSize={10}>
            <div className="flex-column flex-grow-1  gap-2">
              <div className="flex align-items-center justify-content-center pb-3 pl-2 pr-2 w-full">
                <Button
                  outlined
                  className="p-0 border-1 border-solid border-round-3xl border-teal-100 w-full"
                  size="small"
                >
                  <InputText
                    width={100}
                    type="text"
                    className=" border-none"
                    placeholder="Search Docker Images . . . ."
                    // onFocus={handleFocus}
                    // onBlur={handleBlur}
                    size="small"
                  />
                  <i className="pi pi-search text-teal-400 text-2xl pr-1"></i>
                </Button>
              </div>

              <div className="w-full">
                <ImageTagsTreeView nodes={imageTagsTree} height={600}  showImageRepositoryView={showImageRepositoryView} showImageTagView={showImageTagView}/>
              </div>
            </div>
          </SplitterPanel>
          <SplitterPanel minSize={60}>
            {/* <ImagesViewSummary/> */}
            <ImageRepositoryView registery={"proxy-docker-hub"} namespace={"tensorflow"} image_repository={"tensorflow"}/>

          </SplitterPanel>
        </Splitter>
      </div>
    </Sidebar>
  );
};

export default ImagesView;

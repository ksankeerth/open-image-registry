import React, { useState } from "react";
import { TreeNode } from "primereact/treenode";
import { classNames } from "primereact/utils";
import { ImageTagsTreeNodeTypes, RegistryTypes } from "../types/app_types";

export type ImageTagsTreeViewProps = {
  nodes: TreeNode[];
  height: number;
  showImageTagView: (
    registery: string,
    namespace: string,
    imageRepository: string,
    tag: string
  ) => void;
  showImageRepositoryView: (
    registery: string,
    namespace: string,
    imageRepository: string
  ) => void;
};

const ImageTagsTreeView = (props: ImageTagsTreeViewProps) => {
  return (
    <React.Fragment>
      <div
        className=" flex flex-column p-2 border-1 surface-border"
        style={{
          maxHeight: `${props.height}px`,
          overflow: "auto",
          borderRightStyle: "none",
          borderBottomStyle: "none",
          borderLeftStyle: "none",
        }}
      >
        {props.nodes.map((node) => (
          <TreeNodeComponent
            key={node.key}
            node={node}
            showImageRepositoryView={props.showImageRepositoryView}
            showImageTagView={props.showImageTagView}
          />
        ))}
      </div>
    </React.Fragment>
  );
};

export default ImageTagsTreeView;

type TreeNodeComponentProps = {
  node: TreeNode;
  showImageTagView: (
    registery: string,
    namespace: string,
    imageRepository: string,
    tag: string
  ) => void;
  showImageRepositoryView: (
    registery: string,
    namespace: string,
    imageRepository: string
  ) => void;
};

function TreeNodeComponent(props: TreeNodeComponentProps) {
  const [expanded, setExpanded] = useState<Boolean>(
    props.node.expanded == true
  );
  const [selectedPath, setSelectedPath] = useState<string>("");

  const toggleExpanded = () => {
    setExpanded((currentValue) => {
      return !currentValue;
    });
  };

  const handleNodeClick = (nodeType: ImageTagsTreeNodeTypes) => {
    switch (nodeType) {
      case ImageTagsTreeNodeTypes.Repository:
        props.showImageRepositoryView(
          props.node.data.registry,
          props.node.data.namespace,
          props.node.data.image_repository
        );
        break;
      case ImageTagsTreeNodeTypes.Tag:
        props.showImageTagView(
          props.node.data.registry,
          props.node.data.namespace,
          props.node.data.image_repository,
          props.node.data.tag
        );
        break;
    }
  };

  return (
    <div
      className={classNames(
        props.node.leaf ? "flex flex-column" : "flex flex-column gap-0"
      )}
    >
      <div className="flex flex-row gap-0 align-items-center">
        {!props.node.leaf && (
          <div className="cursor-pointer" onClick={toggleExpanded}>
            {" "}
            {expanded ? (
              <span>
                <i className="pi pi-angle-down text-sm"></i>
                {props.node &&
                  props.node.data.node_type !=
                    ImageTagsTreeNodeTypes.Registry && <span>&nbsp;</span>}
              </span>
            ) : (
              <span>
                <i className="pi pi-angle-right text-sm"></i>
                {props.node &&
                  props.node.data.node_type !=
                    ImageTagsTreeNodeTypes.Registry && <span>&nbsp;</span>}
              </span>
            )}
          </div>
        )}

        <div
          onClick={() => {}}
          className={classNames(
            props.node.leaf
              ? "border-none  w-full"
              : "cursor-pointer pt-1 pb-1",
            props.node.data.image_or_tag_path == selectedPath
              ? "surface-100"
              : ""
          )}
        >
          <div
            className={
              props.node.data.node_type == ImageTagsTreeNodeTypes.Tag
                ? "cursor-pointer pr-2 pb-1 underline text-blue-600 text-sm"
                : "cursor-pointer hover:surface-100 pr-2 pl-1"
            }
          >
            {/* {node.data.node_type == ImageTagsTreeNodeTypes.Registry && (
              <span>
                &nbsp;&nbsp;
                {node.data.registry_type == RegistryTypes.Hosted ? (
                  <i className="pi pi-server text-sm " />
                ) : (
                  <i className="pi pi-sync text-sm " />
                )}
              </span>
            )} */}
            {/* &nbsp;&nbsp; */}
            {props.node.label}
          </div>
        </div>
      </div>
      <div className={classNames("flex flex-row p-0")}>
        {!props.node.leaf && expanded && (
          <div
            className="ml-1 mb-3 mr-0"
            style={{
              borderLeft: "0.05rem dashed",
              borderBottom: "0.05rem dashed",
            }}
          ></div>
        )}
        {expanded && props.node.children && props.node.children.length > 0 && (
          <div className="flex flex-column pl-4">
            {props.node.children.map((childNode) => (
              <TreeNodeComponent
                key={childNode.key}
                node={childNode}
                showImageRepositoryView={props.showImageRepositoryView}
                showImageTagView={props.showImageTagView}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

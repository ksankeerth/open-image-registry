export enum RegistryTypes {
  Hosted,
  Proxy,
}

export enum ImageTagsTreeNodeTypes {
  Registry,
  Namespace, 
  Repository,
  Tag,
}

export type ImageTagsTreeViewNodeData = {
  registry_key: string;
  registry_type: RegistryTypes;
  image_or_tag_path: string;
  node_type: ImageTagsTreeNodeTypes;
  namespace: string;
  image_repository: string;
  tag: string;
};


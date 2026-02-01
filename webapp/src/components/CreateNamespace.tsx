import { Button } from 'primereact/button';
import { Checkbox } from 'primereact/checkbox';
import { Dialog } from 'primereact/dialog';
import { Divider } from 'primereact/divider';
import { Dropdown } from 'primereact/dropdown';
import { InputText } from 'primereact/inputtext';
import { InputTextarea } from 'primereact/inputtextarea';
import { classNames } from 'primereact/utils';
import React, { useState } from 'react';
import SearchAndSelectDropdown, { SearchAndSelectDropdownOption } from './SearchAndSelectDropdown';
import { useToast } from './ToastComponent';
import { isValidNamespace } from '../utils';
import { useLoader } from './loader';
import {
  GetUsersData,
  getResourceNamespacesCheckName,
  getUsers,
  postResourceNamespaces,
} from '../api';

export type CreateNamespaceDialogProps = {
  visible: boolean;
  hideCallback: (reload: boolean) => void;
};

const Purposes: Array<'Team' | 'Project'> = ['Project', 'Team'];

const CreateNamespaceDialog = (props: CreateNamespaceDialogProps) => {
  const [name, setName] = useState<string>('');
  const [description, setDescription] = useState<string>('');
  const [selectedMaintainers, setSelectedMaintainers] = useState<SearchAndSelectDropdownOption[]>(
    []
  );
  const [namespacePurpose, setNamespacePurpose] = useState<'Team' | 'Project'>('Project');
  const [isPrivateNamespace, setPrivateNamespace] = useState<boolean>(true);

  const [nameValidationMsg, setNameValidationMsg] = useState<string>('');
  const [purposeValidationMsg, setPurposeValidationMsg] = useState<string>('');
  const [maintainerValidationMsg, setMaintainerValidationMsg] = useState<string>('');
  const [descriptionValidationMsg, setDescriptionValidationMsg] = useState<string>('');

  const [maintainerSuggestions, setMaintainerSuggestions] = useState<
    SearchAndSelectDropdownOption[]
  >([]);
  const [isSearchingMaintainers, setIsSearchingMaintainers] = useState(false);

  const { showSuccess, showError } = useToast();
  const { showLoading, hideLoading } = useLoader();

  const resetFields = () => {
    setName('');
    setDescription('');
    setSelectedMaintainers([]);
    setNamespacePurpose(Purposes[0]);
    setPrivateNamespace(true);
    setNameValidationMsg('');
    setPurposeValidationMsg('');
    setMaintainerValidationMsg('');
    setDescriptionValidationMsg('');
    setMaintainerSuggestions([]);
  };

  const validateNamespaceAvailability = async (name: string, successFn: () => void) => {
    showLoading('Checking name availability ...');
    const { data, error } = await getResourceNamespacesCheckName({
      query: {
        name: name,
      },
    });

    hideLoading();
    if (error) {
      showError(error.error_message);
      return;
    }
    if (!data.available) {
      setNameValidationMsg(`Name ${name} has already taken.`);
    } else {
      setNameValidationMsg('');
    }

    successFn();
  };

  const saveNamespace = async () => {
    showLoading('Creating namespace...');

    const { error } = await postResourceNamespaces({
      body: {
        name,
        description,
        is_public: !isPrivateNamespace,
        purpose: namespacePurpose,
        maintainers: selectedMaintainers.map((v) => v.value),
      },
    });
    hideLoading();

    if (error) {
      showError(error.error_message);
    } else {
      showSuccess(`Successfully created namespace!`);
      resetFields();
    }
  };

  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setName(e.target.value);
    setNameValidationMsg('');
  };

  const handleDescriptionChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setDescription(e.target.value);
    setDescriptionValidationMsg('');
  };

  const handleMaintainersChange = (options: SearchAndSelectDropdownOption[]) => {
    setSelectedMaintainers(options);
    setMaintainerValidationMsg('');
  };

  const searchMaintainers = async (searchTerm: string) => {
    setIsSearchingMaintainers(true);

    const req: GetUsersData = {
      query: {
        limit: 50,
        page: 1,
        sort_by: 'username',
        order: 'asc',
        search: searchTerm,
        role: ['Maintainer'],
        locked: false,
      },
      url: '/users',
    };

    const { data, error } = await getUsers(req);

    setIsSearchingMaintainers(false);

    if (error?.error_message) {
      showError('Failed to search maintainers');
      setMaintainerSuggestions([]);
      return;
    }

    if (data?.entities && data?.entities.length != 0) {
      const options: Array<SearchAndSelectDropdownOption> = data.entities.map((v) => {
        return {
          label: v.username,
          value: v.id,
        };
      });
      setMaintainerSuggestions(options);
    }
  };

  const handleButtonClick = () => {
    let nameMsg = '';
    if (!name) {
      nameMsg = 'Namespace name cannot be empty';
    } else if (!isValidNamespace(name)) {
      nameMsg = 'Invalid namespace name';
    }

    let purposeMsg = '';
    if (!namespacePurpose) {
      purposeMsg = 'Select the purpose';
    }

    let maintainerMsg = '';
    if (selectedMaintainers.length === 0) {
      maintainerMsg = 'Select at least one maintainer';
    }

    let descriptionMsg = '';
    if (!description) {
      descriptionMsg = 'Provide description about the namespace';
    }

    setNameValidationMsg(nameMsg);
    setPurposeValidationMsg(purposeMsg);
    setMaintainerValidationMsg(maintainerMsg);
    setDescriptionValidationMsg(descriptionMsg);

    if (nameMsg || descriptionMsg || purposeMsg || maintainerMsg) {
      return;
    }

    validateNamespaceAvailability(name, saveNamespace);
  };

  return (
    <Dialog
      visible={props.visible}
      onHide={() => {
        resetFields();
        props.hideCallback(false);
      }}
      modal
      dismissableMask
      showHeader={false}
      pt={{
        root: {
          className: 'border-none shadow-none',
        },
        mask: {
          style: {
            backdropFilter: 'blur(8px)',
            backgroundColor: 'rgba(0, 0, 0, 0.3)',
          },
        },
        content: {
          className: 'p-0 m-0 border-round-2xl overflow-hidden',
          style: {
            width: '520px',
            animation: 'dialogSlideIn 0.3s ease-out',
          },
        },
      }}
    >
      <div className="bg-white">
        {/* Header */}
        <div className="px-4 pt-5 pb-2">
          <div className="flex align-items-start justify-content-between mb-1">
            <h2 className="m-0 text-2xl font-semibold text-gray-900">Create Namespace</h2>
            <Button
              icon="pi pi-times"
              onClick={() => {
                resetFields();
                props.hideCallback(false);
              }}
              text
              rounded
              className="w-2rem h-2rem -mt-1 -mr-2"
              pt={{
                root: {
                  className: 'text-gray-400 hover:text-gray-600 hover:bg-gray-50',
                },
              }}
            />
          </div>
          <p className="m-0 text-sm text-gray-500">
            Create a new namespace for your projects or team
          </p>
        </div>

        <Divider layout="horizontal" className="pt-0 mt-0" />

        {/* Form Fields */}
        <div className="px-4 pb-4" style={{ maxHeight: '60vh', overflowY: 'auto' }}>
          <div className="flex flex-column" style={{ gap: '1.25rem' }}>
            {/* Namespace Name */}
            <div className="flex flex-column" style={{ gap: '0.5rem' }}>
              <label htmlFor="namespace-name" className="text-sm font-medium text-gray-700">
                Namespace Name <span className="text-red-500">*</span>
              </label>
              <InputText
                id="namespace-name"
                value={name}
                onChange={handleNameChange}
                placeholder="my-namespace"
                className={classNames('w-full border-round-3xl', {
                  'p-invalid': nameValidationMsg,
                })}
                pt={{
                  root: {
                    className:
                      'text-sm p-3 py-2 border-1 border-gray-200 hover:border-gray-300 focus:border-teal-500',
                    style: {
                      transition: 'all 0.2s ease',
                      boxShadow: nameValidationMsg ? '0 0 0 1px #ef4444' : 'none',
                    },
                  },
                }}
              />
              {nameValidationMsg && (
                <small className="text-red-500 text-xs">{nameValidationMsg}</small>
              )}
            </div>

            {/* Purpose and Privacy */}
            <div className="grid" style={{ margin: 0 }}>
              <div className="col-6 pl-0">
                <div className="flex flex-column" style={{ gap: '0.5rem' }}>
                  <label htmlFor="purpose" className="text-sm font-medium text-gray-700">
                    Purpose <span className="text-red-500">*</span>
                  </label>
                  <Dropdown
                    id="purpose"
                    value={namespacePurpose}
                    onChange={(e) => {
                      setNamespacePurpose(e.value);
                      setPurposeValidationMsg('');
                    }}
                    options={Purposes}
                    placeholder="Select purpose"
                    className={classNames('w-full border-round-3xl', {
                      'p-invalid': purposeValidationMsg,
                    })}
                    pt={{
                      root: {
                        className: 'border-1 border-gray-200 hover:border-gray-300',
                        style: {
                          transition: 'all 0.2s ease',
                          boxShadow: purposeValidationMsg ? '0 0 0 1px #ef4444' : 'none',
                        },
                      },
                      input: {
                        className: 'text-sm',
                      },
                      item: {
                        className: 'text-sm px-3 py-2',
                      },
                      panel: {
                        className: 'border-1 border-gray-200 shadow-lg',
                      },
                    }}
                  />
                  {purposeValidationMsg && (
                    <small className="text-red-500 text-xs">{purposeValidationMsg}</small>
                  )}
                </div>
              </div>

              <div className="col-6 pr-0">
                <div className="flex flex-column pl-3" style={{ gap: '0.5rem' }}>
                  <span className="text-sm font-medium text-gray-700">Visibility</span>
                  <div
                    className="flex flex-grow-1 h-full align-items-center border-none border-gray-200 border-round-3xl px-2 py-2 hover:border-gray-300"
                    style={{
                      transition: 'all 0.2s ease',
                      cursor: 'pointer',
                    }}
                  >
                    <Checkbox
                      inputId="private-namespace"
                      checked={isPrivateNamespace}
                      onChange={(e) => setPrivateNamespace(Boolean(e.checked))}
                      className="mr-2"
                    />
                    <label
                      htmlFor="private-namespace"
                      className="text-sm cursor-pointer"
                      style={{ userSelect: 'none' }}
                    >
                      Private
                    </label>
                  </div>
                </div>
              </div>
            </div>

            {/* Maintainers */}
            <div className="flex flex-column" style={{ gap: '0.5rem' }}>
              <label htmlFor="maintainers" className="text-sm font-medium text-gray-700">
                Maintainers <span className="text-red-500">*</span>
              </label>
              <SearchAndSelectDropdown
                value={selectedMaintainers}
                suggestions={maintainerSuggestions}
                onSelect={handleMaintainersChange}
                onSearch={searchMaintainers}
                placeholder="Search and select maintainers..."
                maxChipsPerRow={3}
                loading={isSearchingMaintainers}
                noResultPlaceholder="No Maintainers found!"
                className={classNames({
                  'p-invalid': maintainerValidationMsg,
                })}
              />
              {maintainerValidationMsg && (
                <small className="text-red-500 text-xs">{maintainerValidationMsg}</small>
              )}
              <small className="text-xs text-gray-400 pl-1">
                Select users who can manage this namespace
              </small>
            </div>

            {/* Description */}
            <div className="flex flex-column" style={{ gap: '0.5rem' }}>
              <label htmlFor="description" className="text-sm font-medium text-gray-700">
                Description <span className="text-red-500">*</span>
              </label>
              <InputTextarea
                id="description"
                value={description}
                onChange={handleDescriptionChange}
                rows={4}
                placeholder="Describe the purpose of this namespace..."
                className={classNames('w-full border-round-lg', {
                  'p-invalid': descriptionValidationMsg,
                })}
                pt={{
                  root: {
                    className:
                      'text-sm p-3 border-1 border-gray-200 hover:border-gray-300 focus:border-teal-500',
                    style: {
                      transition: 'all 0.2s ease',
                      boxShadow: descriptionValidationMsg ? '0 0 0 1px #ef4444' : 'none',
                      resize: 'vertical',
                    },
                  },
                }}
              />
              {descriptionValidationMsg && (
                <small className="text-red-500 text-xs">{descriptionValidationMsg}</small>
              )}
            </div>
          </div>
        </div>

        {/* Footer with Actions */}
        <div className="px-4 py-3 bg-gray-50 flex justify-content-end" style={{ gap: '0.75rem' }}>
          <Button
            label="Cancel"
            onClick={() => {
              resetFields();
              props.hideCallback(false);
            }}
            text
            className="text-sm px-4 py-2 border-round-3xl"
            pt={{
              root: {
                className: 'text-gray-600 hover:bg-gray-100',
                style: {
                  transition: 'all 0.2s ease',
                },
              },
            }}
          />
          <Button
            label="Create Namespace"
            icon="pi pi-check"
            iconPos="left"
            onClick={handleButtonClick}
            className="text-sm px-4 py-2 border-round-3xl"
            pt={{
              root: {
                className: 'hover:shadow-2',
                style: {
                  transition: 'all 0.2s ease',
                },
              },
            }}
          />
        </div>
      </div>
    </Dialog>
  );
};

export default CreateNamespaceDialog;

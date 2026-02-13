import { Button } from 'primereact/button';
import { Dialog } from 'primereact/dialog';
import { Divider } from 'primereact/divider';
import { InputText } from 'primereact/inputtext';
import { InputTextarea } from 'primereact/inputtextarea';
import { classNames } from 'primereact/utils';
import React, { useState } from 'react';
import { useLoader } from './loader';
import { useToast } from './ToastComponent';
import { getResourceRepositoriesCheckName, postResourceRepositories } from '../api';
import { isValidRepository } from '../utils';

export type CreateRepositoryDialogProps = {
  namespaceId: string;
  visible: boolean;
  hideCallback: (reload: boolean) => void;
};

const CreateRepositoryDialog = (props: CreateRepositoryDialogProps) => {
  const [name, setName] = useState<string>('');
  const [description, setDescription] = useState<string>('');
  const [isPublic, setPublic] = useState<boolean>(false);

  const [nameValidationMsg, setNameValidationMsg] = useState<string>('');
  const [descriptionValidationMsg, setDescriptionValidationMsg] = useState<string>('');

  const { showSuccess, showError } = useToast();
  const { showLoading, hideLoading } = useLoader();

  const resetFields = () => {
    setName('');
    setDescription('');
    setPublic(false);
    setNameValidationMsg('');
    setDescriptionValidationMsg('');
  };

  const saveRepository = async () => {
    if (!props.namespaceId) return;

    showLoading('Creating repository...');

    const { error } = await postResourceRepositories({
      body: {
        name,
        description,
        namespace_id: props.namespaceId,
        is_public: isPublic,
      },
    });

    hideLoading();

    if (error) {
      showError(error.error_message);
    } else {
      showSuccess('Successfully created repository!');
      resetFields();
      props.hideCallback(true);
    }
  };

  const validateRepositoryAvailability = async (
    name: string,
    nsId: string,
    successFn: () => void
  ) => {
    showLoading('Checking name availability...');

    const { data, error } = await getResourceRepositoriesCheckName({
      query: {
        name,
        namespace_id: nsId,
      },
    });

    hideLoading();

    if (error) {
      showError(error.error_message);
      return;
    }

    if (!data.available) {
      setNameValidationMsg(`Name "${name}" is already taken.`);
    } else {
      setNameValidationMsg('');
      successFn();
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

  const handleButtonClick = () => {
    let nameMsg = '';
    if (!name) {
      nameMsg = 'Repository name cannot be empty';
    } else if (!isValidRepository(name)) {
      nameMsg = 'Invalid repository name';
    }

    let descriptionMsg = '';
    if (!description) {
      descriptionMsg = 'Provide a description for this repository';
    }

    setNameValidationMsg(nameMsg);
    setDescriptionValidationMsg(descriptionMsg);

    if (nameMsg || descriptionMsg) return;

    validateRepositoryAvailability(name, props.namespaceId, saveRepository);
  };

  const handleHide = () => {
    resetFields();
    props.hideCallback(false);
  };

  return (
    <Dialog
      visible={props.visible}
      onHide={handleHide}
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
            width: '480px',
            animation: 'dialogSlideIn 0.3s ease-out',
          },
        },
      }}
    >
      <div className="bg-white">
        {/* Header */}
        <div className="px-4 pt-5 pb-2">
          <div className="flex align-items-start justify-content-between mb-1">
            <h2 className="m-0 text-2xl font-semibold text-gray-900">Create Repository</h2>
            <Button
              icon="pi pi-times"
              onClick={handleHide}
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
            Create a new repository to store your container images
          </p>
        </div>

        <Divider layout="horizontal" className="pt-0 mt-0" />

        {/* Form */}
        <div className="px-4 pb-4">
          <div className="flex flex-column" style={{ gap: '1.25rem' }}>
            {/* Repository Name */}
            <div className="flex flex-column" style={{ gap: '0.5rem' }}>
              <label htmlFor="repo-name" className="text-sm font-medium text-gray-700">
                Repository Name <span className="text-red-500">*</span>
              </label>
              <InputText
                id="repo-name"
                value={name}
                onChange={handleNameChange}
                placeholder="my-repository"
                className={classNames('w-full border-round-3xl', {
                  'p-invalid': nameValidationMsg,
                })}
                pt={{
                  root: {
                    className:
                      'text-sm py-2 px-3 border-1 border-gray-200 hover:border-gray-300 focus:border-teal-500',
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
              <small className="text-xs text-gray-400 pl-1">
                Only lowercase letters, numbers, and hyphens allowed
              </small>
            </div>

            {/* Visibility */}
            <div className="flex flex-column" style={{ gap: '0.5rem' }}>
              <span className="text-sm font-medium text-gray-700">Visibility</span>
              <div className="flex flex-row gap-4">
                {/* Private Option */}
                <div
                  className={classNames(
                    'flex flex-1 align-items-center gap-2 px-3 py-2 border-1 border-round-3xl cursor-pointer transition-all transition-duration-200',
                    !isPublic
                      ? 'border-teal-400 bg-teal-50'
                      : 'border-gray-200 hover:border-gray-300'
                  )}
                  onClick={() => setPublic(false)}
                >
                  <i
                    className={classNames(
                      'pi pi-lock text-sm',
                      !isPublic ? 'text-teal-500' : 'text-gray-400'
                    )}
                  />
                  <div className="flex flex-column" style={{ gap: '0.1rem' }}>
                    <span
                      className={classNames(
                        'text-sm font-medium',
                        !isPublic ? 'text-teal-700' : 'text-gray-700'
                      )}
                    >
                      Private
                    </span>
                    <span className="text-xs text-gray-400">Only authorized users</span>
                  </div>
                </div>

                {/* Public Option */}
                <div
                  className={classNames(
                    'flex flex-1 align-items-center gap-2 px-3 py-2 border-1 border-round-3xl cursor-pointer transition-all transition-duration-200',
                    isPublic
                      ? 'border-teal-400 bg-teal-50'
                      : 'border-gray-200 hover:border-gray-300'
                  )}
                  onClick={() => setPublic(true)}
                >
                  <i
                    className={classNames(
                      'pi pi-globe text-sm',
                      isPublic ? 'text-teal-500' : 'text-gray-400'
                    )}
                  />
                  <div className="flex flex-column" style={{ gap: '0.1rem' }}>
                    <span
                      className={classNames(
                        'text-sm font-medium',
                        isPublic ? 'text-teal-700' : 'text-gray-700'
                      )}
                    >
                      Public
                    </span>
                    <span className="text-xs text-gray-400">Anyone can pull images</span>
                  </div>
                </div>
              </div>
            </div>

            {/* Description */}
            <div className="flex flex-column" style={{ gap: '0.5rem' }}>
              <label htmlFor="repo-description" className="text-sm font-medium text-gray-700">
                Description <span className="text-red-500">*</span>
              </label>
              <InputTextarea
                id="repo-description"
                value={description}
                onChange={handleDescriptionChange}
                rows={3}
                placeholder="Describe what this repository contains..."
                className={classNames('w-full border-round-lg', {
                  'p-invalid': descriptionValidationMsg,
                })}
                pt={{
                  root: {
                    className:
                      'text-sm p-3 border-1 border-gray-200 hover:border-gray-300 focus:border-teal-500',
                    style: {
                      transition: 'all 0.2s ease',
                      resize: 'vertical',
                      boxShadow: descriptionValidationMsg ? '0 0 0 1px #ef4444' : 'none',
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

        {/* Footer */}
        <div className="px-4 py-3 bg-gray-50 flex justify-content-end" style={{ gap: '0.75rem' }}>
          <Button
            label="Cancel"
            onClick={handleHide}
            text
            className="text-sm px-4 py-2 border-round-3xl"
            pt={{
              root: {
                className: 'text-gray-600 hover:bg-gray-100',
                style: { transition: 'all 0.2s ease' },
              },
            }}
          />
          <Button
            label="Create Repository"
            icon="pi pi-check"
            iconPos="left"
            onClick={handleButtonClick}
            className="text-sm px-4 py-2 border-round-3xl"
            pt={{
              root: {
                className: 'hover:shadow-2',
                style: { transition: 'all 0.2s ease' },
              },
            }}
          />
        </div>
      </div>
    </Dialog>
  );
};

export default CreateRepositoryDialog;

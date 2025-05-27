import React, { useState } from "react";
import { Outlet } from "react-router-dom";
import logo from "./../assets/logo.png";
import { Sidebar } from "primereact/sidebar";
import { DataTable } from "primereact/datatable";
import { Column } from "primereact/column";
import { Stepper } from "primereact/stepper";
import { StepperPanel } from "primereact/stepperpanel";
import { Divider } from "primereact/divider";
import { Fieldset } from "primereact/fieldset";
import { Button } from "primereact/button";
import UpstreamRegistry from "./UpstreamsRegistry";
import ImagesView from "./ImagesView";
import LogoComponent from "./LogoComponent";

const RootLayout = () => {
  const [showUpstreamModal, setShowUpstreamModal] = useState<boolean>(false);
  const [showImagesModal, setShowImagesModal] = useState<boolean>(false);

  return (
    <div className="flex flex-column min-h-screen max-h-screen">
      <div className="flex-grow-0  h-5rem w-screen flex flex-row justify-content-betweenjustify-content-end align-items-center mr-4 gap-3">
        <LogoComponent/>

        <div className="flex-grow-1 flex justify-content-end gap-3 pr-4 text-color">
          <div
            className="cursor-pointer"
            style={{ zIndex: 50 }}
            onClick={() => setShowImagesModal(true)}
          >
            Images
          </div>
          <div
            className="cursor-pointer"
            style={{ zIndex: 50 }}
            onClick={() => setShowUpstreamModal(true)}
          >
            Upstreams
          </div>
          <div className="cursor-pointer" style={{ zIndex: 50 }}>
            Settings
          </div>
        </div>
      </div>
      <div className="flex-grow-1 flex align-items-stretch">
        <Outlet />
      </div>
      {/* For upstreams */}
      <UpstreamRegistry
        visible={showUpstreamModal}
        hideCallback={setShowUpstreamModal}
      />

      {/* For Images view */}
      <ImagesView visible={showImagesModal} hideCallback={setShowImagesModal} />
    </div>
  );
};

export default RootLayout;

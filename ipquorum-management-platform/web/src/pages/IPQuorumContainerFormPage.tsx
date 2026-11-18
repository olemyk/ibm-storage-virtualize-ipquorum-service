import React from 'react';
import { useNavigate } from 'react-router-dom';
import { IPQuorumContainerForm } from '../components/IPQuorumContainerForm';
import { useToastStore } from '../stores/toastStore';

const IPQuorumContainerFormPage: React.FC = () => {
  const navigate = useNavigate();
  const addToast = useToastStore((state) => state.addToast);

  const handleSuccess = (containerId: string) => {
    addToast({
      type: 'success',
      message: `IPQuorum container created successfully!`,
    });
    
    // Navigate directly to containers page (no delay needed)
    navigate('/containers');
  };

  const handleCancel = () => {
    navigate('/ipquorum-containers');
  };

  return (
    <IPQuorumContainerForm
      onSuccess={handleSuccess}
      onCancel={handleCancel}
    />
  );
};

export default IPQuorumContainerFormPage;

// Made with Bob
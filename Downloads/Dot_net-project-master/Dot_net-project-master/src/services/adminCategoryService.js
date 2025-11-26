// src/services/adminCategoryService.js
import apiClient from './api';

export const getAllAdminCategories = async () => {
  try {
    const response = await apiClient.get('/categories'); // Using public endpoint for listing
    return response.data;
  } catch (error) {
    console.error("Error fetching categories for admin:", error.response?.data || error.message);
    throw error;
  }
};

export const getCategoryByIdAdmin = async (categoryId) => {
  try {
    // Assuming the public GET /api/categories/{id} is sufficient
    const response = await apiClient.get(`/categories/${categoryId}`);
    return response.data; // Expects CategoryDto { id, name }
  } catch (error) {
    console.error(`Error fetching category ${categoryId}:`, error.response?.data || error.message);
    throw error;
  }
};

export const createCategory = async (categoryData) => {
  // categoryData should be { name: "string" } matching CreateCategoryDto
  try {
    const response = await apiClient.post('/admin/categories', categoryData);
    return response.data; // Expects the created CategoryDto
  } catch (error) {
    console.error("Error creating category:", error.response?.data || error.message);
    throw error;
  }
};

export const updateCategory = async (categoryId, categoryData) => {
  // categoryData should be { name: "string" } matching UpdateCategoryDto
  try {
    const response = await apiClient.put(`/admin/categories/${categoryId}`, categoryData);
    return response.data; // Or just status for success
  } catch (error) {
    console.error(`Error updating category ${categoryId}:`, error.response?.data || error.message);
    throw error;
  }
};

export const deleteCategory = async (categoryId) => {
  try {
    const response = await apiClient.delete(`/admin/categories/${categoryId}`);
    return response.data;
  } catch (error) {
    console.error(`Error deleting category ${categoryId}:`, error.response?.data || error.message);
    throw error;
  }
};
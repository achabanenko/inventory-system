import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Toaster } from 'react-hot-toast';
import { AuthProvider } from './contexts/AuthContext';
import ProtectedRoute from './components/ProtectedRoute';
import Layout from './components/Layout';

import Login from './pages/Login';
import Register from './pages/Register';
import TenantLookup from './pages/TenantLookup';
import Dashboard from './pages/Dashboard';
import Items from './pages/Items';
import Inventory from './pages/Inventory';
import PurchaseOrders from './pages/PurchaseOrders';
import Transfers from './pages/Transfers';
import Adjustments from './pages/Adjustments';
import Suppliers from './pages/Suppliers';
import Locations from './pages/Locations';
import Categories from './pages/Categories';
import Users from './pages/Users';
import Counts from './pages/Counts';
import Receipts from './pages/Receipts';
import GoodsReceiptDetails from './components/GoodsReceiptDetails';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5,
      retry: 1,
    },
  },
});

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <Router>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route path="/tenant-lookup" element={<TenantLookup />} />
            <Route
              path="/"
              element={
                <ProtectedRoute>
                  <Layout />
                </ProtectedRoute>
              }
            >
              <Route index element={<Navigate to="/dashboard" replace />} />
              <Route path="dashboard" element={<Dashboard />} />
              <Route path="items" element={<Items />} />
              <Route path="inventory" element={<Inventory />} />
              <Route path="purchase-orders" element={<PurchaseOrders />} />
              <Route path="transfers" element={<Transfers />} />
              <Route path="adjustments" element={<Adjustments />} />
              <Route path="counts" element={<Counts />} />
              <Route path="receipts" element={<Receipts />} />
              <Route path="receipts/:id" element={<GoodsReceiptDetails />} />
              <Route path="suppliers" element={<Suppliers />} />
              <Route path="locations" element={<Locations />} />
              <Route path="categories" element={<Categories />} />
              <Route path="users" element={<Users />} />
            </Route>
          </Routes>
        </Router>
        <Toaster />
      </AuthProvider>
    </QueryClientProvider>
  );
}

export default App;
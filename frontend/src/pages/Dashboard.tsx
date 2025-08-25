import { Package, TrendingUp, ShoppingCart, AlertTriangle } from 'lucide-react';

const stats = [
  { name: 'Total Items', value: '2,451', icon: Package, change: '+12%' },
  { name: 'Low Stock Items', value: '23', icon: AlertTriangle, change: '-5%' },
  { name: 'Pending Orders', value: '12', icon: ShoppingCart, change: '+3%' },
  { name: 'Total Value', value: '$245,091', icon: TrendingUp, change: '+18%' },
];

export default function Dashboard() {
  return (
    <div>
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p className="mt-1 text-sm text-gray-500">
          Overview of your inventory management system
        </p>
      </div>

      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
        {stats.map((stat) => (
          <div key={stat.name} className="bg-white overflow-hidden shadow rounded-lg">
            <div className="p-5">
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <stat.icon className="h-6 w-6 text-gray-400" />
                </div>
                <div className="ml-5 w-0 flex-1">
                  <dl>
                    <dt className="text-sm font-medium text-gray-500 truncate">
                      {stat.name}
                    </dt>
                    <dd className="flex items-baseline">
                      <div className="text-2xl font-semibold text-gray-900">
                        {stat.value}
                      </div>
                      <div className="ml-2 flex items-baseline text-sm font-semibold text-green-600">
                        {stat.change}
                      </div>
                    </dd>
                  </dl>
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="mt-8 grid grid-cols-1 gap-6 lg:grid-cols-2">
        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-lg font-medium text-gray-900 mb-4">Recent Activities</h2>
          <div className="space-y-3">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="flex items-start space-x-3">
                <div className="flex-shrink-0">
                  <div className="h-8 w-8 rounded-full bg-gray-100 flex items-center justify-center">
                    <Package className="h-4 w-4 text-gray-500" />
                  </div>
                </div>
                <div className="flex-1">
                  <p className="text-sm text-gray-900">
                    Purchase Order #PO-2024-{String(i).padStart(3, '0')} received
                  </p>
                  <p className="text-xs text-gray-500">2 hours ago</p>
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-lg font-medium text-gray-900 mb-4">Low Stock Alerts</h2>
          <div className="space-y-3">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-900">Item SKU-{String(i).padStart(4, '0')}</p>
                  <p className="text-xs text-gray-500">Current stock: {i * 2} units</p>
                </div>
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800">
                  Low Stock
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
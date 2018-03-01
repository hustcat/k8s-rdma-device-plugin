package ibverbs

type IbvDevice struct {
	/* Name of underlying kernel IB device, eg "mlx5_0" */
	Name string

	/* Name of uverbs device, eg "uverbs0" */
	DevName string

	/* Path to infiniband_verbs class device in sysfs */
	DevPath string

	/* Path to infiniband class device in sysfs */
	IbvDevPath string
}

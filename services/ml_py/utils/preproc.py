import numpy as np


def uniform_sample(arr, N=75):
    """Uniformly sample N frames from the input array.

    Parameters
    ----------
    arr : array-like
        Sequence of frames. The first dimension represents time.
    N : int, default 75
        Number of samples to return.

    Returns
    -------
    numpy.ndarray
        Array sampled to length N along the first axis.

    Raises
    ------
    ValueError
        If the input array is empty.
    """
    a = np.asarray(arr)
    if a.size == 0:
        raise ValueError("input array is empty")
    indices = np.linspace(0, len(a) - 1, num=N).astype(int)
    return a[indices]

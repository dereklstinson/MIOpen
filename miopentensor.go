package miopen

/*
#include <miopen/miopen.h>

*/
import "C"
import (
	"runtime"

	"github.com/dereklstinson/cutil"
)

const miopendimmax = 5

//TensorD is a tensor descriptor
type TensorD struct {
	d    C.miopenTensorDescriptor_t
	gogc bool
	dims C.int
}

//CreateTensorDescriptor creates an empty tensor descriptor
func CreateTensorDescriptor() (*TensorD, error) {
	return createtensordescriptor()

}
func destroytensordescriptor(t *TensorD) error {
	return Status(C.miopenDestroyTensorDescriptor(t.d)).error("destroytensordescriptor")

}
func createtensordescriptor() (*TensorD, error) {
	d := new(TensorD)

	err := Status(C.miopenCreateTensorDescriptor(&d.d)).error("CreateTensorDescriptor-create")
	if err != nil {
		return nil, err
	}

	runtime.SetFinalizer(d, destroytensordescriptor)

	return d, nil
}

//Set sets the t's values
func (t *TensorD) Set(data DataType, shape, stride []int32) error {
	t.dims = (C.int)(len(shape))
	//	t.dtype = data
	if stride == nil {
		cstride := stridecalc(shape)
		stridecint := int32Tocint(cstride)
		shapecint := int32Tocint(shape)
		return Status(C.miopenSetTensorDescriptor(t.d, data.c(), t.dims, &shapecint[0], &stridecint[0])).error("cudnnSetTensorNdDescriptorEx-set")
	}

	shapecint := int32Tocint(shape)
	stridecint := int32Tocint(stride)
	return Status(C.miopenSetTensorDescriptor(t.d, data.c(), t.dims, &shapecint[0], &stridecint[0])).error("cudnnSetTensorNdDescriptor")

}

//Get gets t's values
//
//Get the details of the N-dimensional tensor descriptor.
func (t *TensorD) Get() (dtype DataType, shape []int32, stride []int32, err error) {
	if t.dims == 0 {
		t.dims = miopendimmax
		shapec := make([]C.int, t.dims)
		stridec := make([]C.int, t.dims)
		err = Status(C.miopenGetTensorDescriptor(t.d, dtype.cptr(), &shapec[0], &stridec[0])).error("cudnnSetTensorNdDescriptor")

		shape = cintToint32(shapec)
		stride = cintToint32(stridec)
		return dtype, shape, stride, err
	}
	shapec := make([]C.int, t.dims)
	stridec := make([]C.int, t.dims)
	err = Status(C.miopenGetTensorDescriptor(t.d, dtype.cptr(), &shapec[0], &stridec[0])).error("cudnnSetTensorNdDescriptor")
	shape = cintToint32(shapec)
	stride = cintToint32(stridec)
	return dtype, shape, stride, err
}

//GetNumOfElements - Get Tensor Volume by elements
//
//Interface for querying tensor size. MIOpen has support for 1, 2, 3, 4, 5 dimensional tensor of layout.
func (t *TensorD) GetNumOfElements() (num int32, err error) {
	err = Status(C.miopenGetTensorDescriptorSize(t.d, (*C.int)(&num))).error("GetNumOfElements")
	return num, err
}

//SetAll - Fills a tensor with a single value.
//
func (t *TensorD) SetAll(h *Handle, tmem cutil.Mem, alpha float64) error {
	dtype, _, _, err := t.Get()
	if err != nil {
		return err
	}
	val := cscalarbydatatype(dtype, alpha)
	return Status(C.miopenSetTensor(h.x, t.d, tmem.Ptr(), val.CPtr())).error("SetAll")

}

//Scale - Scales all elements in a tensor by a single value.
//
//	h		MiOpen handle (input)
//
//	tmem		Tensor Memory (input and output)
//
//	alpha		Floating point scaling factor, allocated on the host (input)
//
func (t *TensorD) Scale(h *Handle, tmem cutil.Mem, alpha float64) error {
	dtype, _, _, err := t.Get()
	if err != nil {
		return err
	}
	val := cscalarbydatatype(dtype, alpha)
	return Status(C.miopenScaleTensor(h.x, t.d, tmem.Ptr(), val.CPtr())).error("Scale")

}

//GetSIB -  Returns number of bytes associated with tensor descriptor
//
func (t *TensorD) GetSIB() (sib uint, err error) {
	var sizet C.size_t
	err = Status(C.miopenGetTensorNumBytes(t.d, &sizet)).error("GetSIB")
	sib = (uint)(sizet)
	return sib, err
}

//TransformTensor - Copies one tensor to another tensor with a different layout.
//
//	h         MIOpen handle (input)
//
//	alpha     Floating point scaling factor, allocated on the host (input)
//
//	xD	      Source Tensor descriptor for tensor x (input)
//
//	x         Source Tensor x (input)
//
//	beta       Floating point scaling factor, allocated on the host (input)
//
//	yD	      Destination Tensor descriptor for tensor y (input)
//
//	y         Destination Tensor y (output)
//
func TransformTensor(h *Handle, alpha float64, xD *TensorD, x cutil.Mem, beta float64, yD *TensorD, y cutil.Mem) error {
	dtype, _, _, err := xD.Get()
	if err != nil {
		return err
	}
	a := cscalarbydatatype(dtype, alpha)
	b := cscalarbydatatype(dtype, beta)
	return Status(C.miopenTransformTensor(h.x, a.CPtr(), xD.d, x.Ptr(), b.CPtr(), yD.d, y.Ptr())).error("TransformTensor")
}

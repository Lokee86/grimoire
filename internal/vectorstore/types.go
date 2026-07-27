package vectorstore

import lodestone "github.com/Lokee86/lodestone/bindings/go/lodestone"

const ABIName = lodestone.ABIName

var (
	ErrUnavailable    = lodestone.ErrUnavailable
	ErrBufferTooSmall = lodestone.ErrBufferTooSmall
	FindLibrary       = lodestone.FindLibrary
	Load              = lodestone.Load
)

type Info = lodestone.Info
type Hit = lodestone.Hit
type Library = lodestone.Library
type Engine = lodestone.Engine

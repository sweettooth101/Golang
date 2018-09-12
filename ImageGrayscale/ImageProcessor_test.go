package main

import "testing"

var expectedResults = map[string]bool{
	"./images/1.jpg":     true,
	"./images/2.jpg":     true,
	"./images/toTheMoon": false,
}

func TestGrayScalePlain(t *testing.T) {

	for k, v := range expectedResults {
		if result := GrayScale(k) == nil; result != v {
			t.Errorf("GrayScale %q returned %t, expected %t",
				k, result, v)
		}
	}

}

func TestGrayScaleRbr(t *testing.T) {

	for k, v := range expectedResults {
		if result := GrayScaleConcurrencyRbr(k, 5) == nil; result != v {
			t.Errorf("GrayScale %q returned %t, expected %t",
				k, result, v)
		}
	}

}

func TestGrayScaleParts(t *testing.T) {

	for k, v := range expectedResults {
		if result := GrayScaleConcurrencyParts(k, 5) == nil; result != v {
			t.Errorf("GrayScale %q returned %t, expected %t",
				k, result, v)
		}
	}

}

func BenchmarkGrayScalePlain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GrayScaleExecute("./images/3.jpg", "./images/ben_plain_grayout.jpg")
	}

}

func BenchmarkGrayScaleRbr2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GrayScaleConcurrencyExcuteRbr("./images/3.jpg", "./images/ben_rbr2_grayout.jpg", 2)
	}

}

func BenchmarkGrayScaleRbr5(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GrayScaleConcurrencyExcuteRbr("./images/3.jpg", "./images/ben_rbr5_grayout.jpg", 5)
	}

}

func BenchmarkGrayScaleParts2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GrayScaleConcurrencyExcuteParts("./images/3.jpg", "./images/ben_part2_grayout.jpg", 2)
	}

}

func BenchmarkGrayScaleParts5(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GrayScaleConcurrencyExcuteParts("./images/3.jpg", "./images/ben_part5_grayout.jpg", 5)
	}

}

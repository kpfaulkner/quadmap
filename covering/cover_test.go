package covering

import (
	"fmt"
	"testing"

	"github.com/kpfaulkner/quadmap/quadtree"
	"github.com/peterstace/simplefeatures/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func quadKeyToWKT(t *testing.T, qk quadtree.QuadKey) string {
	env, err := qk.Envelope()
	require.NoError(t, err)
	return env.AsGeometry().AsText()
}

func TestExternalCovering(t *testing.T) {
	for _, tc := range []struct {
		name     string
		wkt      string
		maxCells int
		expect   []quadtree.QuadKey
	}{
		{
			name:     "whole map",
			wkt:      quadKeyToWKT(t, quadtree.QuadKey(0)),
			maxCells: 20,
			expect: []quadtree.QuadKey{
				quadtree.GenerateQuadKeyIndexFromSlippy(0, 0, 0),
			},
		},
		{
			name:     "single point",
			wkt:      "POINT(151.196 -33.866)",
			maxCells: 20,
			expect: []quadtree.QuadKey{
				quadtree.GenerateQuadKeyIndexFromSlippy(493915273, 322167045, 29),
			},
		},
		{
			name:     "linestring",
			wkt:      "LINESTRING (151.17777883970012 -33.89915886674083, 151.17944876656372 -33.896744381970066, 151.1812802992535 -33.89571596954773, 151.18365051802806 -33.89334610240112, 151.18736745201534 -33.892183501988086, 151.19033022548223 -33.88941108320346, 151.19323913034145 -33.88766709618549, 151.19474745138047 -33.88458149334736)",
			maxCells: 20,
			expect: []quadtree.QuadKey{
				0xd6c7c80fd0000012,
				0xd6c7c80fe0000012,
				0xd6c7c80ff0000012,
				0xd6c7c81a00000012,
				0xd6c7c81a20000012,
				0xd6c7c81a80000012,
				0xd6c7c82340000011,
				0xd6c7c82380000011,
				0xd6c7c823c0000011,
				0xd6c7c824a0000012,
				0xd6c7c824b0000012,
				0xd6c7c824c0000011,
				0xd6c7c82510000012,
				0xd6c7c82520000012,
				0xd6c7c82530000012,
				0xd6c7c82540000012,
				0xd6c7c82580000012,
				0xd6c7c82600000011,
				0xd6c7c82900000011,
			},
		},
		{
			name:     "single cell includes neighbouring cells because they intersect the boundary",
			wkt:      quadKeyToWKT(t, quadtree.GenerateQuadKeyIndexFromSlippy(123, 456, 10)),
			maxCells: 20,
			expect: []quadtree.QuadKey{
				0x2b56ef000000000c,
				0x2b56f8000000000b,
				0x2b56fe000000000c,
				0x2b56ff000000000c,
				0x2b57aa000000000c,
				0x2b5c45000000000c,
				0x2b5c47000000000c,
				0x2b5c4d000000000c,
				0x2b5c4f000000000c,
				0x2b5c50000000000a, // original cell
				0x2b5c65000000000c,
				0x2b5c70000000000c,
				0x2b5c71000000000c,
				0x2b5c74000000000c,
				0x2b5c75000000000c,
				0x2b5d00000000000c,
				0x2b5d02000000000c,
				0x2b5d08000000000c,
				0x2b5d0a000000000c,
				0x2b5d20000000000c,
			},
		},
		{
			name: "Australia",
			wkt: `MULTIPOLYGON(
				((115.12974936961064 -33.94746740383465, 116.89325344621824 -35.1773935246154, 123.54635525699587 -34.0334665647765, 125.01229427555933 -32.76504696519842, 130.8533264250692 -31.621205514074042, 133.3554878149327 -32.013812745097916, 135.59925229667363 -34.824810219542044, 140.15501878018097 -37.94107655667957, 143.29965711410927 -38.98670541298011, 145.99722125772973 -39.15763221685892, 149.89173421681983 -37.7629793608208, 153.28147327805465 -31.274083836016892, 153.19944941207405 -25.699910662918327, 142.40106983051436 -10.445759124437714, 140.49420368110157 -17.547905748173463, 135.53694091389116 -14.833465482045824, 136.88524622800003 -12.169310284547564, 130.56764876913473 -11.235979504388865, 129.25811634656884 -14.111505857716836, 129.55354736823062 -14.99121754617586, 126.93839179015254 -13.866803186408347, 125.14157328273859 -14.493436274092332, 121.13901156170527 -19.316348563404404, 113.74170321256048 -21.997856972782103, 113.47904014429406 -26.171395434387343, 115.84534035714637 -32.53204953697848, 115.12974936961064 -33.94746740383465)),
				((144.28919920677697 -40.77079688015533, 146.07335528591045 -43.71159773845069, 147.43011180353886 -43.616346924564745, 148.33034010300855 -40.908336071447536, 146.25199570018162 -41.090890980452386, 144.28919920677697 -40.77079688015533)))`,
			maxCells: 20,
			expect: []quadtree.QuadKey{
				0xd6c0000000000005,
				0xd640000000000005,
				0xd430000000000006,
				0xd480000000000005,
				0xd680000000000005,
				0xdc40000000000006,
				0xd4c0000000000006,
				0xd190000000000006,
				0xd3d0000000000006,
				0xd3c0000000000006,
				0xd1b0000000000006,
				0xd600000000000005,
				0xd340000000000005,
				0xd1c0000000000005,
				0xd4e0000000000006,
				0xdc10000000000006,
				0xd1a0000000000006,
				0xd300000000000005,
				0xd380000000000006,
				0xd390000000000006,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, err := geom.UnmarshalWKT(tc.wkt)
			require.NoError(t, err)

			cov, err := ExteriorCovering(g, tc.maxCells)
			require.NoError(t, err)
			assert.ElementsMatch(t, tc.expect, cov)
		})
	}
}

func TestSearchRanges(t *testing.T) {
	for _, tc := range []struct {
		name    string
		cells   []quadtree.QuadKey
		minZoom byte
		expect  []quadtree.QuadKeyRange
	}{
		{
			name:    "whole map",
			cells:   []quadtree.QuadKey{quadtree.QuadKey(0)},
			minZoom: 0,
			expect: []quadtree.QuadKeyRange{
				{Start: 0x0000000000000000, End: 0xffffffffffffffff},
			},
		},
		{
			name:    "single tile",
			cells:   []quadtree.QuadKey{quadtree.GenerateQuadKeyIndexFromSlippy(123, 456, 9)},
			minZoom: 0,
			expect: []quadtree.QuadKeyRange{
				{Start: 0x0000000000000000, End: 0x0000000000000000},
				{Start: 0x8000000000000001, End: 0x8000000000000001},
				{Start: 0xa000000000000002, End: 0xa000000000000002},
				{Start: 0xac00000000000003, End: 0xac00000000000003},
				{Start: 0xad00000000000004, End: 0xad00000000000004},
				{Start: 0xad40000000000005, End: 0xad40000000000005},
				{Start: 0xad70000000000006, End: 0xad70000000000007},
				{Start: 0xad71000000000008, End: 0xad71000000000008},
				{Start: 0xad71400000000000, End: 0xad717fffffffffff},
			},
		},
		{
			name:    "single tile with minZoom",
			cells:   []quadtree.QuadKey{quadtree.GenerateQuadKeyIndexFromSlippy(123, 456, 9)},
			minZoom: 5,
			expect: []quadtree.QuadKeyRange{
				{Start: 0xad40000000000005, End: 0xad40000000000005},
				{Start: 0xad70000000000006, End: 0xad70000000000007},
				{Start: 0xad71000000000008, End: 0xad71000000000008},
				{Start: 0xad71400000000000, End: 0xad717fffffffffff},
			},
		},
		{
			name: "multiple tiles, some contiguous",
			cells: []quadtree.QuadKey{
				0xd300000000000004,
				0xd600000000000004,
				0xd180000000000005,
				0xd400000000000005,
				0xdc40000000000005,
				0xdc00000000000005,
				0xd4c0000000000005,
			},
			minZoom: 2,
			expect: []quadtree.QuadKeyRange{
				{Start: 0xd000000000000002, End: 0xd000000000000003},
				{Start: 0xd100000000000004, End: 0xd100000000000004},
				{Start: 0xd180000000000000, End: 0xd1bfffffffffffff},
				{Start: 0xd300000000000000, End: 0xd43fffffffffffff},
				{Start: 0xd4c0000000000000, End: 0xd4ffffffffffffff},
				{Start: 0xd600000000000000, End: 0xd6ffffffffffffff},
				{Start: 0xdc00000000000000, End: 0xdc7fffffffffffff},
			},
		},
		{
			name: "multiple tiles, minZoom > tile zoom, some contiguous",
			cells: []quadtree.QuadKey{
				0xd400000000000004,
				0xd100000000000004,
				0xd300000000000004,
				0xd600000000000004,
				0xdc00000000000004,
			},
			minZoom: 5,
			expect: []quadtree.QuadKeyRange{
				{Start: 0xd100000000000000, End: 0xd1ffffffffffffff},
				{Start: 0xd300000000000000, End: 0xd4ffffffffffffff},
				{Start: 0xd600000000000000, End: 0xd6ffffffffffffff},
				{Start: 0xdc00000000000000, End: 0xdcffffffffffffff},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := SearchRanges(tc.cells, tc.minZoom)
			require.NoError(t, err)
			fmt.Println(tc.name)
			for _, qk := range r {
				fmt.Printf("{Start: 0x%016x, End: 0x%016x},\n", qk.Start, qk.End)
			}
			assert.Equal(t, tc.expect, r)
		})
	}
}

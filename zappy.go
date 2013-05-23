// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Copyright 2011 The Snappy-Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the SNAPPY-GO-LICENSE file.

/*
Package zappy implements the zappy block-based compression format.  It aims for
a combination of good speed and reasonable compression.

Zappy is a format incompatible, API compatible fork of snappy-go[1]. The C++
snappy implementation is at [2].

Reasons for the fork

The snappy compression is pretty good. Yet it has one problem built into its
format definition[3] - the maximum length of a copy "instruction" is 64 bytes.
For some specific usage patterns with long runs of repeated data, it turns out
the compression is suboptimal. For example a 1:1000 "sparseness" 64kB bit index
with only few set bits is compressed to about 3kB (about 1000 of 64B copy, 3
byte "instructions").

Format description

Zappy uses much less complicated format than snappy. Each encoded block begins
with the uvarint-encoded[4] length of the decoded data, followed by a sequence
of chunks. Chunks begin and end on byte boundaries. The chunk starts with a
varint encoded number N:

	N >= 0: N+1 literal bytes follow.

	N < 0: copy -N bytes, starting at offset M (in the following uvarint).

Performance issues

Compression rate is roughly the same as of snappy for the reference data set:

                  testdata/html: snappy   23320, zappy   22943, 0.984, orig  102400
              testdata/urls.10K: snappy  334437, zappy  355163, 1.062, orig  702087
             testdata/house.jpg: snappy  126711, zappy  126694, 1.000, orig  126958
  testdata/mapreduce-osdi-1.pdf: snappy   77227, zappy   77646, 1.005, orig   94330
              testdata/html_x_4: snappy   92350, zappy   22956, 0.249, orig  409600
               testdata/cp.html: snappy   11938, zappy   12961, 1.086, orig   24603
              testdata/fields.c: snappy    4825, zappy    5395, 1.118, orig   11150
           testdata/grammar.lsp: snappy    1814, zappy    1933, 1.066, orig    3721
           testdata/kennedy.xls: snappy  423518, zappy  440597, 1.040, orig 1029744
           testdata/alice29.txt: snappy   89550, zappy  104016, 1.162, orig  152089
          testdata/asyoulik.txt: snappy   79583, zappy   91345, 1.148, orig  125179
            testdata/lcet10.txt: snappy  238761, zappy  275488, 1.154, orig  426754
          testdata/plrabn12.txt: snappy  324567, zappy  376885, 1.161, orig  481861
                  testdata/ptt5: snappy   96350, zappy   91465, 0.949, orig  513216
                   testdata/sum: snappy   18927, zappy   20015, 1.057, orig   38240
               testdata/xargs.1: snappy    2532, zappy    2793, 1.103, orig    4227
         testdata/geo.protodata: snappy   23362, zappy   20759, 0.889, orig  118588
             testdata/kppkn.gtb: snappy   73962, zappy   87200, 1.179, orig  184320
                          TOTAL: snappy 2043734, zappy 2136254, 1.045, orig 4549067

Zappy has better RLE handling (1/1000+1 non zero bytes in each index):

     Sparse bit index      16 B: snappy       9, zappy       9, 1.000
     Sparse bit index      32 B: snappy      10, zappy      10, 1.000
     Sparse bit index      64 B: snappy      11, zappy      10, 0.909
     Sparse bit index     128 B: snappy      16, zappy      14, 0.875
     Sparse bit index     256 B: snappy      22, zappy      14, 0.636
     Sparse bit index     512 B: snappy      36, zappy      16, 0.444
     Sparse bit index    1024 B: snappy      57, zappy      18, 0.316
     Sparse bit index    2048 B: snappy     111, zappy      32, 0.288
     Sparse bit index    4096 B: snappy     210, zappy      31, 0.148
     Sparse bit index    8192 B: snappy     419, zappy      75, 0.179
     Sparse bit index   16384 B: snappy     821, zappy     138, 0.168
     Sparse bit index   32768 B: snappy    1627, zappy     232, 0.143
     Sparse bit index   65536 B: snappy    3243, zappy     451, 0.139

When compiled with CGO_ENABLED=1, zappy is now faster than snappy-go.
Old=snappy-go, new=zappy:

 benchmark                   old MB/s     new MB/s  speedup
 BenchmarkWordsDecode1e3       148.98       189.04    1.27x
 BenchmarkWordsDecode1e4       150.29       182.51    1.21x
 BenchmarkWordsDecode1e5       145.79       182.95    1.25x
 BenchmarkWordsDecode1e6       167.43       187.69    1.12x
 BenchmarkWordsEncode1e3        47.11       145.69    3.09x
 BenchmarkWordsEncode1e4        81.47       136.50    1.68x
 BenchmarkWordsEncode1e5        78.86       127.93    1.62x
 BenchmarkWordsEncode1e6        96.81       142.95    1.48x
 Benchmark_UFlat0              316.87       463.19    1.46x
 Benchmark_UFlat1              231.56       350.32    1.51x
 Benchmark_UFlat2             3656.68      8258.39    2.26x
 Benchmark_UFlat3              892.56      1270.09    1.42x
 Benchmark_UFlat4              315.84       959.08    3.04x
 Benchmark_UFlat5              211.70       301.55    1.42x
 Benchmark_UFlat6              211.59       258.29    1.22x
 Benchmark_UFlat7              209.80       272.21    1.30x
 Benchmark_UFlat8              254.59       301.70    1.19x
 Benchmark_UFlat9              163.39       192.66    1.18x
 Benchmark_UFlat10             155.46       189.70    1.22x
 Benchmark_UFlat11             170.11       198.95    1.17x
 Benchmark_UFlat12             148.32       178.78    1.21x
 Benchmark_UFlat13             359.25       579.99    1.61x
 Benchmark_UFlat14             197.27       291.33    1.48x
 Benchmark_UFlat15             185.75       248.07    1.34x
 Benchmark_UFlat16             362.74       582.66    1.61x
 Benchmark_UFlat17             222.95       240.01    1.08x
 Benchmark_ZFlat0              188.66       311.89    1.65x
 Benchmark_ZFlat1              101.46       201.34    1.98x
 Benchmark_ZFlat2               93.62       244.50    2.61x
 Benchmark_ZFlat3              102.79       243.34    2.37x
 Benchmark_ZFlat4              191.64       625.32    3.26x
 Benchmark_ZFlat5              103.09       169.39    1.64x
 Benchmark_ZFlat6              110.35       182.57    1.65x
 Benchmark_ZFlat7               89.56       190.53    2.13x
 Benchmark_ZFlat8              154.05       235.68    1.53x
 Benchmark_ZFlat9               87.58       133.51    1.52x
 Benchmark_ZFlat10              82.08       127.51    1.55x
 Benchmark_ZFlat11              91.36       138.91    1.52x
 Benchmark_ZFlat12              79.24       123.02    1.55x
 Benchmark_ZFlat13             217.04       374.26    1.72x
 Benchmark_ZFlat14             100.33       168.03    1.67x
 Benchmark_ZFlat15              80.79       160.46    1.99x
 Benchmark_ZFlat16             213.32       375.79    1.76x
 Benchmark_ZFlat17             135.37       197.13    1.46x

The package buils with CGO_ENABLED=0 as well, but the performance is worse.

Information sources

... referenced from the above documentation.

 [1]: http://code.google.com/p/snappy-go/
 [2]: http://code.google.com/p/snappy/
 [3]: http://code.google.com/p/snappy/source/browse/trunk/format_description.txt
 [4]: http://golang.org/pkg/encoding/binary/
*/
package zappy

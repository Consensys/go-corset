;;error:4:22-23:incompatible length multiplier
(defcolumns X Y)
(definterleaved A (X Y))
(definterleaved B (X A))

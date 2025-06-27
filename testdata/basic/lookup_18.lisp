(defcolumns (X :i16) (Y :i16) (Z :i16))
;; fragmented lookup with two targets
(defmlookup test ((Y) (Z)) ((X)))

(defpurefun ((i1 :i1 :force) x) x)
(defcolumns (X :i8) (Y :i16) (Z :i1@prove) (P :i2))

(defconstraint c1 ()
  (if (== P 1)
      (begin
       (== Z X)
       (== 130 (+ 123 (i1 X) Y)))))

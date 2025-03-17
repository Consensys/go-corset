(defpurefun (vanishes! x) (== 0 x))

(defcolumns (X :i16) (Y :i16) (Z :i16))
(defconstraint c1 ()
  (if (== 0 X)
      (begin
       (vanishes! Y)
       (vanishes! Z))))

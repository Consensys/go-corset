(defpurefun (vanishes! x) (== 0 x))

(defcolumns (X :i16) (Y :i16) (Z :i16))
(defconstraint test ()
  (if (== 0 X)
      (vanishes! 0)
      (== Z (if (== 0 Y) 3 16))))

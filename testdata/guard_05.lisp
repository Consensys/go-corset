(defpurefun (vanishes! x) (== 0 x))

(defcolumns (ST :i4) (X :i16) (Y :i16) (Z :i16))
(defconstraint test (:guard ST)
  (if (== 0 X)
      (vanishes! (- Z (if (== 0 Y) 0 16)))))

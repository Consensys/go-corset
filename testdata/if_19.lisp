(defpurefun ((not! :bool) (x :bool)) (if x (!= 0 0) (== 0 0)))

(defcolumns (X :i16) (Y :i16) (Z :i16))

(defconstraint c1 ()
  (if (not! (== X Y))
          (== 0 Z)))

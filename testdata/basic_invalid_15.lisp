;;error:6:18-21:expected bool, found u1
(defpurefun (and! (a :bool) (b :bool)) (if a b (!= 0 0)))
(defcolumns (X :i16) (BIT :binary))

(defconstraint c1 ()
  (and! (== X 0) BIT))

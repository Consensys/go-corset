(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (X :i16) (Y :i16))
(defconstraint c1 ()
  (begin
   (vanishes! X)
   (vanishes! Y)))

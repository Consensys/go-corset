(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (ST :i16) (A :i16))
(defconstraint spills ()
  (vanishes!
   (* ST (* A (~ (shift A 2))))))

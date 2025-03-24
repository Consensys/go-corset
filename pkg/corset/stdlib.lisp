;; Copyright Consensys Software Inc.
;;
;; Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
;; the License. You may obtain a copy of the License at
;;
;; http://www.apache.org/licenses/LICENSE-2.0
;;
;; Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
;; an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
;; specific language governing permissions and limitations under the License.
;;
;; SPDX-License-Identifier: Apache-2.0
(defpurefun (if-zero cond then) (if (vanishes! cond) then))
(defpurefun (if-zero cond then else) (if (vanishes! cond) then else))

(defpurefun (if-not-zero cond then) (if (as-bool cond) then))
(defpurefun (if-not-zero cond then else) (if (as-bool cond) then else))

(defpurefun ((as-bool :𝔽@bool :force) x) x)
(defpurefun ((is-binary :𝔽@loob :force) e0) (* e0 (- 1 e0)))

(defpurefun ((force-bin :binary :force) x) x)

;;
;; Boolean functions
;;
;; !-suffix denotes loobean algebra (i.e. 0 == true)
;; ~-prefix denotes normalized-functions (i.e. output is 0/1)
(defpurefun (and a b) (* a b))
(defpurefun ((or! :𝔽@loob) (a :𝔽@loob) (b :𝔽@loob)) (* a b))
(defpurefun ((or! :𝔽@loob) (a :𝔽@loob) (b :𝔽@loob) (c :𝔽@loob)) (* a b c))
(defpurefun ((not :binary@bool :force) (x :binary)) (- 1 x))

(defpurefun ((eq! :𝔽@loob) x y) (- x y))
(defpurefun ((neq! :binary@loob :force) x y) (not (~ (eq! x y))))

(defpurefun ((eq :binary@bool :force) x y) (- 1 (~ (eq! x y))))
(defpurefun ((neq :𝔽@bool) x y) (- x y))

;; Boolean functions
(defpurefun ((is-not-zero :binary@bool) x) (~ x))
(defpurefun ((is-not-zero! :binary@loob :force) x) (- 1 (is-not-zero x)))
(defpurefun ((is-zero :binary@bool :force) x) (- 1 (~ x)))

;; Chronological functions
(defpurefun (next X) (shift X 1))
(defpurefun (prev X) (shift X -1))

;; Ensure that e0 has (resp. will) increase (resp. decrease) of offset
;; w.r.t. the previous (resp. next) row.
(defpurefun (did-inc! e0 offset) (eq! e0 (+ (prev e0) offset)))
(defpurefun (did-dec! e0 offset) (eq!  e0 (- (prev e0) offset)))
(defpurefun (will-inc! e0 offset) (will-eq! e0 (+ e0 offset)))
(defpurefun (will-dec! e0 offset) (eq! (next e0) (- e0 offset)))

(defpurefun (did-inc e0 offset) (eq e0 (+ (prev e0) offset)))
(defpurefun (did-dec e0 offset) (eq  e0 (- (prev e0) offset)))
(defpurefun (will-inc e0 offset) (will-eq e0 (+ e0 offset)))
(defpurefun (will-dec e0 offset) (eq (next e0) (- e0 offset)))

;; Ensure that e0 remained (resp. will be) constant
;; with regards to the previous (resp. next) row.
(defpurefun (remained-constant! e0) (eq! e0 (prev e0)))
(defpurefun (will-remain-constant! e0) (will-eq! e0 e0))

(defpurefun (remained-constant e0) (eq e0 (prev e0)))
(defpurefun (will-remain-constant e0) (will-eq e0 e0))

;; Ensure (in loobean logic) that e0 has changed (resp. will change) its value
;; with regards to the previous (resp. next) row.
(defpurefun (did-change! e0) (neq! e0 (prev e0)))
(defpurefun (will-change! e0) (neq! e0 (next e0)))

(defpurefun (did-change e0) (neq e0 (prev e0)))
(defpurefun (will-change e0) (neq e0 (next e0)))

;; Ensure (in loobean logic) that e0 was (resp. will be) equal to e1 in the
;; previous (resp. next) row.
(defpurefun (was-eq! e0 e1) (eq! (prev e0) e1))
(defpurefun (will-eq! e0 e1) (eq! (next e0) e1))

(defpurefun (was-eq e0 e1) (eq (prev e0) e1))
(defpurefun (will-eq e0 e1) (eq (next e0) e1))

;; Helpers
(defpurefun ((vanishes! :𝔽@loob :force) e0) e0)
(defpurefun (if-eq x val then) (if (eq! x val) then))
(defpurefun (if-eq-else x val then else) (if (eq! x val) then else))
(defpurefun (if-not-eq A B then) (if (neq A B) then))

;; counter constancy constraint
(defpurefun ((counter-constancy :𝔽@loob) ct X)
  (if-not-zero ct
               (remained-constant! X)))

;; perspective constancy constraint
(defpurefun ((perspective-constancy :𝔽@loob) PERSPECTIVE_SELECTOR X)
            (if-not-zero (* PERSPECTIVE_SELECTOR (prev PERSPECTIVE_SELECTOR))
                         (remained-constant! X)))

;; base-X decomposition constraints
(defpurefun (base-X-decomposition ct base acc digits)
  (if-zero ct
           (eq! acc digits)
           (eq! acc (+ (* base (prev acc)) digits))))

;; byte decomposition constraint
(defpurefun (byte-decomposition ct acc bytes) (base-X-decomposition ct 256 acc bytes))

;; bit decomposition constraint
(defpurefun (bit-decomposition ct acc bits) (base-X-decomposition ct 2 acc bits))

;; plateau constraints
(defpurefun (plateau-constraint CT (X :binary) C)
            (begin (debug (stamp-constancy CT C))
                   (if-zero C
                            (eq! X 1)
                            (if (eq! CT 0)
                                (vanishes! X)
                              (if (eq!  CT C)
                                  (did-inc! X 1)
                                (remained-constant! X))))))

;; stamp constancy imposes that the column C may only
;; change at rows where the STAMP column changes.
(defpurefun (stamp-constancy STAMP C)
            (if (will-remain-constant! STAMP)
                (will-remain-constant! C)))
